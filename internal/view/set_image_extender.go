package view

import (
	"context"
	"fmt"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
)

// SetImageExtender adds set image extensions
type SetImageExtender struct {
	ResourceViewer
}

type ContainerType string

type ContainerImage struct {
	ContainerType ContainerType
	Image         string
}

const (
	setImageKey      = "setImage"
	runningContainer = ContainerType("Container")
	initContainer    = ContainerType("InitContainer")
)

func NewSetImageExtender(r ResourceViewer) ResourceViewer {
	s := SetImageExtender{ResourceViewer: r}
	s.bindKeys(s.Actions())

	return &s
}

func (s *SetImageExtender) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyI: ui.NewKeyAction("SetImage", s.setImageCmd, true),
	})
}

func (s *SetImageExtender) setImageCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return nil
	}

	s.Stop()
	defer s.Start()
	s.showSetImageDialog(path)

	return nil
}

func (s *SetImageExtender) showSetImageDialog(path string) {
	confirm := tview.NewModalForm("<Set image>", s.makeSetImageForm(path))
	confirm.SetText(fmt.Sprintf("Set image %s %s", s.GVR(), path))
	confirm.SetDoneFunc(func(int, string) {
		s.dismissDialog()
	})
	s.App().Content.AddPage(setImageKey, confirm, false, false)
	s.App().Content.ShowPage(setImageKey)
}

func (s *SetImageExtender) makeSetImageForm(sel string) *tview.Form {
	f := s.makeStyledForm()
	podSpec, err := s.getPodSpec(sel)
	originalImages := getImages(podSpec)
	formSubmitResult := make(map[string]ContainerImage, 0)
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}
	for name, containerImage := range originalImages {
		f.AddInputField(name, containerImage.Image, 0, nil, func(changed string) {
			formSubmitResult[name] = ContainerImage{ContainerType: containerImage.ContainerType, Image: changed}
		})
	}

	f.AddButton("OK", func() {
		defer s.dismissDialog()

		if err != nil {
			s.App().Flash().Err(err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), s.App().Conn().Config().CallTimeout())
		defer cancel()
		podSpecPatch := buildPodSpecPatch(formSubmitResult, originalImages)
		if err := s.setImages(ctx, sel, podSpecPatch); err != nil {

			log.Error().Err(err).Msgf("PodSpec %s image update failed", sel)
			s.App().Flash().Err(err)
		} else {
			s.App().Flash().Infof("Resource %s:%s image updated successfully", s.GVR(), sel)
		}
	})
	f.AddButton("Cancel", func() {
		s.dismissDialog()
	})

	return f
}

func getImages(podSpec *corev1.PodSpec) map[string]ContainerImage {
	results := make(map[string]ContainerImage, 0)
	for _, c := range podSpec.Containers {
		results[c.Name] = ContainerImage{
			ContainerType: runningContainer,
			Image:         c.Image,
		}
	}
	for _, c := range podSpec.InitContainers {
		results[c.Name] = ContainerImage{
			ContainerType: initContainer,
			Image:         c.Image,
		}
	}
	return results
}

func buildPodSpecPatch(formImages map[string]ContainerImage, originalImages map[string]ContainerImage) corev1.PodSpec {
	initContainers := make([]corev1.Container, 0)
	containers := make([]corev1.Container, 0)
	for name, containerImage := range formImages {
		if originalImages[name].Image == containerImage.Image {
			continue
		}
		container := corev1.Container{
			Image: containerImage.Image,
			Name:  name,
		}
		switch containerImage.ContainerType {
		case runningContainer:
			containers = append(containers, container)
		case initContainer:
			initContainers = append(initContainers, container)
		}
	}
	result := corev1.PodSpec{
		Containers:     containers,
		InitContainers: initContainers,
	}
	return result
}

func (s *SetImageExtender) dismissDialog() {
	s.App().Content.RemovePage(setImageKey)
}

func (s *SetImageExtender) makeStyledForm() *tview.Form {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldTextColor(tcell.ColorOrange)
	return f
}

func (s *SetImageExtender) getPodSpec(path string) (*corev1.PodSpec, error) {
	res, err := dao.AccessorFor(s.App().factory, s.GVR())
	if err != nil {
		return nil, err
	}
	resourceWPodSpec, ok := res.(dao.ContainsPodSpec)
	if !ok {
		return nil, fmt.Errorf("expecting a resourceWPodSpec resource for %q", s.GVR())
	}
	podSpec, err := resourceWPodSpec.GetPodSpec(path)
	return podSpec, nil
}

func (s *SetImageExtender) setImages(ctx context.Context, path string, spec corev1.PodSpec) error {
	res, err := dao.AccessorFor(s.App().factory, s.GVR())
	if err != nil {
		return err
	}

	resourceWPodSpec, ok := res.(dao.ContainsPodSpec)
	if !ok {
		return fmt.Errorf("expecting a scalable resource for %q", s.GVR())
	}

	return resourceWPodSpec.SetImages(ctx, path, spec)
}
