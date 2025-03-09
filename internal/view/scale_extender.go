// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/ui"
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

	meta, err := dao.MetaAccess.MetaFor(s.GVR())
	if err != nil {
		slog.Error("No meta information found",
			slogs.GVR, s.GVR(),
			slogs.Error, err,
		)
		return
	}

	if !dao.IsCRD(meta) || dao.IsScalable(meta) {
		aa.Add(ui.KeyS, ui.NewKeyActionWithOpts("Scale", s.scaleCmd,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			},
		))
	}
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

func (s *ScaleExtender) replicasFromReady(_ string) (string, error) {
	replicas, err := s.valueOf("READY")
	if err != nil {
		return "", err
	}

	tokens := strings.Split(replicas, "/")
	if len(tokens) < 2 {
		return "", fmt.Errorf("unable to locate replicas from %s", replicas)
	}

	return strings.TrimRight(tokens[1], ui.DeltaSign), nil
}

func (s *ScaleExtender) replicasFromScaleSubresource(sel string) (string, error) {
	res, err := dao.AccessorFor(s.App().factory, s.GVR())
	if err != nil {
		return "", err
	}

	replicasGetter, ok := res.(dao.ReplicasGetter)
	if !ok {
		return "", fmt.Errorf("expecting a replicasGetter resource for %q", s.GVR())
	}

	ctx, cancel := context.WithTimeout(context.Background(), s.App().Conn().Config().CallTimeout())
	defer cancel()

	replicas, err := replicasGetter.Replicas(ctx, sel)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(int(replicas)), nil
}

func (s *ScaleExtender) makeScaleForm(fqns []string) (*tview.Form, error) {
	factor := "0"
	if len(fqns) == 1 {
		// If the CRD resource supports scaling, then first try to
		// read the replicas directly from the CRD.
		if meta, _ := dao.MetaAccess.MetaFor(s.GVR()); dao.IsScalable(meta) {
			replicas, err := s.replicasFromScaleSubresource(fqns[0])
			if err == nil && len(replicas) != 0 {
				factor = replicas
			}
		}

		// For built-in resources or cases where we can't get the replicas from the CRD, we can
		// only try to get the number of copies from the READY field.
		if factor == "0" {
			replicas, err := s.replicasFromReady(fqns[0])
			if err != nil {
				return nil, err
			}

			factor = replicas
		}
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
		for _, fqn := range fqns {
			if err := s.scale(ctx, fqn, count); err != nil {
				slog.Error("Unable to scale resource", slogs.FQN, fqn)
				s.App().Flash().Err(err)
				return
			}
		}
		if len(fqns) != 1 {
			s.App().Flash().Infof("[%d] %s scaled successfully", len(fqns), singularize(s.GVR().R()))
		} else {
			s.App().Flash().Infof("%s %s scaled successfully", s.GVR().R(), fqns[0])
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
