package view

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

type ScaleExtender struct {
	ResourceViewer
}

func NewScaleExtender(r ResourceViewer) ResourceViewer {
	s := ScaleExtender{ResourceViewer: r}
	s.BindKeys()

	return &s
}

func (s *ScaleExtender) BindKeys() {
	s.Actions().Add(ui.KeyActions{
		ui.KeyS: ui.NewKeyAction("Scale", s.scaleCmd, true),
	})
}

func (s *ScaleExtender) scaleCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return nil
	}

	s.Stop()
	defer s.Start()
	s.showScaleDialog(path)

	return nil
}

func (s *ScaleExtender) showScaleDialog(path string) {
	confirm := tview.NewModalForm("<Scale>", s.makeScaleForm(path))
	confirm.SetText(fmt.Sprintf("Scale %s %s", s.List().GetName(), path))
	confirm.SetDoneFunc(func(int, string) {
		s.dismissDialog()
	})
	s.App().Content.AddPage(scaleDialogKey, confirm, false, false)
	s.App().Content.ShowPage(scaleDialogKey)
}

func (s *ScaleExtender) makeScaleForm(path string) *tview.Form {
	f := s.makeStyledForm()
	replicas := strings.TrimSpace(s.GetTable().GetCell(s.GetTable().GetSelectedRowIndex(), s.GetTable().NameColIndex()+1).Text)
	f.AddInputField("Replicas:", replicas, 4, func(textToCheck string, lastChar rune) bool {
		_, err := strconv.Atoi(textToCheck)
		return err == nil
	}, func(changed string) {
		replicas = changed
	})

	f.AddButton("OK", func() {
		defer s.dismissDialog()
		count, err := strconv.Atoi(replicas)
		if err != nil {
			s.App().Flash().Err(err)
			return
		}
		if err := s.scale(path, count); err != nil {
			s.App().Flash().Err(err)
		} else {
			s.App().Flash().Infof("Resource %s:%s scaled successfully", s.List().GetName(), path)
		}
	})

	f.AddButton("Cancel", func() {
		s.dismissDialog()
	})

	return f
}

func (s *ScaleExtender) dismissDialog() {
	s.App().Content.RemovePage(scaleDialogKey)
}

func (s *ScaleExtender) makeStyledForm() *tview.Form {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldTextColor(tcell.ColorOrange)

	return f
}

func (s *ScaleExtender) scale(path string, replicas int) error {
	ns, n := namespaced(path)
	scaler, ok := s.List().Resource().(resource.Scalable)
	if !ok {
		return errors.New("Expecting a valid scalable resource")
	}

	return scaler.Scale(ns, n, int32(replicas))
}
