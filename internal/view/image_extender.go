// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
	corev1 "k8s.io/api/core/v1"
)

const imageKey = "setImage"

type imageFormSpec struct {
	name, dockerImage, newDockerImage string
	init                              bool
}

func (m *imageFormSpec) modified() bool {
	newDockerImage := strings.TrimSpace(m.newDockerImage)
	return newDockerImage != "" && m.dockerImage != newDockerImage
}

func (m *imageFormSpec) imageSpec() dao.ImageSpec {
	ret := dao.ImageSpec{
		Name: m.name,
		Init: m.init,
	}

	if m.modified() {
		ret.DockerImage = strings.TrimSpace(m.newDockerImage)
	} else {
		ret.DockerImage = m.dockerImage
	}

	return ret
}

// ImageExtender provides for overriding container images.
type ImageExtender struct {
	ResourceViewer
}

// NewImageExtender returns a new extender.
func NewImageExtender(r ResourceViewer) ResourceViewer {
	s := ImageExtender{ResourceViewer: r}
	s.AddBindKeysFn(s.bindKeys)

	return &s
}

func (s *ImageExtender) bindKeys(aa *ui.KeyActions) {
	if s.App().Config.K9s.IsReadOnly() {
		return
	}
	aa.Add(ui.KeyI, ui.NewKeyAction("Set Image", s.setImageCmd, false))
}

func (s *ImageExtender) setImageCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return nil
	}

	s.Stop()
	defer s.Start()
	if err := s.showImageDialog(path); err != nil {
		s.App().Flash().Err(err)
	}

	return nil
}

func (s *ImageExtender) showImageDialog(path string) error {
	form, err := s.makeSetImageForm(path)
	if err != nil {
		return err
	}
	confirm := tview.NewModalForm("<Set image>", form)
	confirm.SetText(fmt.Sprintf("Set image %s %s", s.GVR(), path))
	confirm.SetDoneFunc(func(int, string) {
		s.dismissDialog()
	})
	s.App().Content.AddPage(imageKey, confirm, false, false)
	s.App().Content.ShowPage(imageKey)

	return nil
}

func (s *ImageExtender) makeSetImageForm(sel string) (*tview.Form, error) {
	podSpec, err := s.getPodSpec(sel)
	if err != nil {
		return nil, err
	}

	formContainerLines := make([]*imageFormSpec, 0, len(podSpec.InitContainers)+len(podSpec.Containers))
	for _, spec := range podSpec.InitContainers {
		formContainerLines = append(formContainerLines, &imageFormSpec{init: true, name: spec.Name, dockerImage: spec.Image})
	}
	for _, spec := range podSpec.Containers {
		formContainerLines = append(formContainerLines, &imageFormSpec{name: spec.Name, dockerImage: spec.Image})
	}

	styles := s.App().Styles.Dialog()
	f := tview.NewForm().
		SetItemPadding(0).
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color()).
		AddButton("OK", func() {
			defer s.dismissDialog()
			var imageSpecsModified dao.ImageSpecs
			for _, v := range formContainerLines {
				if v.modified() {
					imageSpecsModified = append(imageSpecsModified, v.imageSpec())
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), s.App().Conn().Config().CallTimeout())
			defer cancel()
			if err := s.setImages(ctx, sel, imageSpecsModified); err != nil {
				log.Error().Err(err).Msgf("PodSpec %s image update failed", sel)
				s.App().Flash().Err(err)
				return
			}
			s.App().Flash().Infof("Resource %s:%s image updated successfully", s.GVR(), sel)
		}).
		AddButton("Cancel", func() {
			s.dismissDialog()
		})

	for i := range formContainerLines {
		ctn := formContainerLines[i]
		f.AddInputField(ctn.name, ctn.dockerImage, 0, nil, func(changed string) {
			ctn.newDockerImage = changed
		})
	}

	for i := 0; i < f.GetButtonCount(); i++ {
		f.GetButton(i).
			SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color()).
			SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
	}

	return f, nil
}

func (s *ImageExtender) dismissDialog() {
	s.App().Content.RemovePage(imageKey)
}

func (s *ImageExtender) getPodSpec(path string) (*corev1.PodSpec, error) {
	res, err := dao.AccessorFor(s.App().factory, s.GVR())
	if err != nil {
		return nil, err
	}
	resourceWPodSpec, ok := res.(dao.ContainsPodSpec)
	if !ok {
		return nil, fmt.Errorf("expecting a ContainsPodSpec for %q but got %T", s.GVR(), res)
	}

	return resourceWPodSpec.GetPodSpec(path)
}

func (s *ImageExtender) setImages(ctx context.Context, path string, imageSpecs dao.ImageSpecs) error {
	res, err := dao.AccessorFor(s.App().factory, s.GVR())
	if err != nil {
		return err
	}

	resourceWPodSpec, ok := res.(dao.ContainsPodSpec)
	if !ok {
		return fmt.Errorf("expecting a scalable resource for %q", s.GVR())
	}

	return resourceWPodSpec.SetImages(ctx, path, imageSpecs)
}
