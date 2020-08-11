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

const setImageKey = "setImage"

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
	containers, err := s.getImages(sel)
	if err != nil {
		s.App().Flash().Err(err)
		return nil
	}
	images := make(map[string]string)
	for _, container := range *containers {
		f.AddInputField(container.Name, container.Image, 0, nil, func(changed string) {
			images[container.Name] = changed
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
		if err := s.setImages(ctx, sel, images); err != nil {
			log.Error().Err(err).Msgf("DP %s image update failed", sel)
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

func (s *SetImageExtender) getImages(path string) (*[]corev1.Container, error) {
	res, err := dao.AccessorFor(s.App().factory, s.GVR())
	if err != nil {
		return nil, err
	}
	podSpecable, ok := res.(dao.ContainsPodSpec)
	if !ok {
		return nil, fmt.Errorf("expecting a podSpecable resource for %q", s.GVR())
	}
	podSpec, err := podSpecable.GetPodSpec(path)
	containers := podSpec.Containers
	return &containers, nil
}

func (s *SetImageExtender) setImages(ctx context.Context, path string, images map[string]string) error {
	res, err := dao.AccessorFor(s.App().factory, s.GVR())
	if err != nil {
		return err
	}

	deployment, ok := res.(dao.ContainsPodSpec)
	if !ok {
		return fmt.Errorf("expecting a scalable resource for %q", s.GVR())
	}

	return deployment.SetImages(ctx, path, images)
}
