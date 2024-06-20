// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
)

// ScaleExtender adds scaling extensions.
type ScaleExtender struct {
	ResourceViewer
}

// NewScaleExtender returns a new extender.
func NewScaleExtender(r ResourceViewer) ResourceViewer {
	s := ScaleExtender{ResourceViewer: r}
	s.AddBindKeysFn(s.bindKeys)

	return &s
}

func (s *ScaleExtender) bindKeys(aa *ui.KeyActions) {
	if s.App().Config.K9s.IsReadOnly() {
		return
	}
	aa.Add(ui.KeyS, ui.NewKeyActionWithOpts("Scale", s.scaleCmd,
		ui.ActionOpts{
			Visible:   true,
			Dangerous: true,
		},
	))
}

func (s *ScaleExtender) scaleCmd(evt *tcell.EventKey) *tcell.EventKey {
	paths := s.GetTable().GetSelectedItems()
	if len(paths) == 0 {
		return nil
	}

	s.Stop()
	defer s.Start()
	s.showScaleDialog(paths)

	return nil
}

func (s *ScaleExtender) showScaleDialog(paths []string) {
	form, err := s.makeScaleForm(paths)
	if err != nil {
		s.App().Flash().Err(err)
		return
	}
	confirm := tview.NewModalForm("<Scale>", form)
	msg := fmt.Sprintf("Scale %s %s?", singularize(s.GVR().R()), paths[0])
	if len(paths) > 1 {
		msg = fmt.Sprintf("Scale [%d] %s?", len(paths), s.GVR().R())
	}
	confirm.SetText(msg)
	confirm.SetDoneFunc(func(int, string) {
		s.dismissDialog()
	})
	s.App().Content.AddPage(scaleDialogKey, confirm, false, false)
	s.App().Content.ShowPage(scaleDialogKey)
}

func (s *ScaleExtender) valueOf(col string) (string, error) {
	colIdx, ok := s.GetTable().HeaderIndex(col)
	if !ok {
		return "", fmt.Errorf("no column index for %s", col)
	}
	return s.GetTable().GetSelectedCell(colIdx), nil
}

func (s *ScaleExtender) makeScaleForm(sels []string) (*tview.Form, error) {
	factor := "0"
	if len(sels) == 1 {
		replicas, err := s.valueOf("READY")
		if err != nil {
			return nil, err
		}
		tokens := strings.Split(replicas, "/")
		if len(tokens) < 2 {
			return nil, fmt.Errorf("unable to locate replicas from %s", replicas)
		}
		factor = strings.TrimRight(tokens[1], ui.DeltaSign)
	}

	styles := s.App().Styles.Dialog()
	f := tview.NewForm().
		SetItemPadding(0).
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())

	f.AddInputField("Replicas:", factor, 4, func(textToCheck string, lastChar rune) bool {
		_, err := strconv.Atoi(textToCheck)
		return err == nil
	}, func(changed string) {
		factor = changed
	})

	f.AddButton("OK", func() {
		defer s.dismissDialog()
		count, err := strconv.Atoi(factor)
		if err != nil {
			s.App().Flash().Err(err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), s.App().Conn().Config().CallTimeout())
		defer cancel()
		for _, sel := range sels {
			if err := s.scale(ctx, sel, count); err != nil {
				log.Error().Err(err).Msgf("DP %s scaling failed", sel)
				s.App().Flash().Err(err)
				return
			}
		}
		if len(sels) == 1 {
			s.App().Flash().Infof("[%d] %s scaled successfully", len(sels), singularize(s.GVR().R()))
		} else {
			s.App().Flash().Infof("%s %s scaled successfully", s.GVR().R(), sels[0])
		}
	})
	f.AddButton("Cancel", func() {
		s.dismissDialog()
	})
	for i := 0; i < 2; i++ {
		if b := f.GetButton(i); b != nil {
			b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
			b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
		}
	}

	for i := 0; i < f.GetButtonCount(); i++ {
		f.GetButton(i).
			SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color()).
			SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
	}

	return f, nil
}

func (s *ScaleExtender) dismissDialog() {
	s.App().Content.RemovePage(scaleDialogKey)
}

func (s *ScaleExtender) scale(ctx context.Context, path string, replicas int) error {
	res, err := dao.AccessorFor(s.App().factory, s.GVR())
	if err != nil {
		return err
	}
	scaler, ok := res.(dao.Scalable)
	if !ok {
		return fmt.Errorf("expecting a scalable resource for %q", s.GVR())
	}

	return scaler.Scale(ctx, path, int32(replicas))
}
