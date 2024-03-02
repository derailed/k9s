// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render/helm"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
)

// History represents a helm History view.
type History struct {
	ResourceViewer

	Values *model.RevValues
}

// NewHistory returns a new helm-history view.
func NewHistory(gvr client.GVR) ResourceViewer {
	h := History{
		ResourceViewer: NewValueExtender(NewBrowser(gvr)),
	}
	h.GetTable().SetColorerFn(helm.History{}.ColorerFunc())
	h.GetTable().SetBorderFocusColor(tcell.ColorMediumSpringGreen)
	h.GetTable().SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorMediumSpringGreen).Attributes(tcell.AttrNone))
	h.AddBindKeysFn(h.bindKeys)
	h.SetContextFn(h.HistoryContext)
	h.GetTable().SetEnterFn(h.getValsCmd)

	return &h
}

// Init initializes the view
func (h *History) Init(ctx context.Context) error {
	if err := h.ResourceViewer.Init(ctx); err != nil {
		return err
	}
	h.GetTable().SetSortCol("REVISION", false)

	return nil
}

func (h *History) HistoryContext(ctx context.Context) context.Context {
	return ctx
}

func (h *History) bindKeys(aa *ui.KeyActions) {
	if !h.App().Config.K9s.IsReadOnly() {
		h.bindDangerousKeys(aa)
	}

	aa.Delete(ui.KeyShiftA, ui.KeyShiftN, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace, tcell.KeyCtrlD)
	aa.Bulk(ui.KeyMap{
		ui.KeyShiftN: ui.NewKeyAction("Sort Revision", h.GetTable().SortColCmd("REVISION", true), false),
		ui.KeyShiftS: ui.NewKeyAction("Sort Status", h.GetTable().SortColCmd("STATUS", true), false),
		ui.KeyShiftA: ui.NewKeyAction("Sort Age", h.GetTable().SortColCmd("AGE", true), false),
	})
}

func (h *History) getValsCmd(app *App, _ ui.Tabular, _ client.GVR, path string) {
	ns, n := client.Namespaced(path)
	tt := strings.Split(n, ":")
	if len(tt) < 2 {
		app.Flash().Err(fmt.Errorf("unable to parse version in %q", path))
		return
	}
	name, rev := tt[0], tt[1]
	h.Values = model.NewRevValues(h.GVR(), client.FQN(ns, name), rev)
	v := NewLiveView(h.App(), "Values", h.Values)
	if err := v.app.inject(v, false); err != nil {
		v.app.Flash().Err(err)
	}
}

func (h *History) bindDangerousKeys(aa *ui.KeyActions) {
	aa.Add(ui.KeyR, ui.NewKeyActionWithOpts("RollBackTo...", h.rollbackCmd,
		ui.ActionOpts{
			Visible:   true,
			Dangerous: true,
		},
	))
}

func (h *History) rollbackCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := h.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	ns, nrev := client.Namespaced(path)
	tt := strings.Split(nrev, ":")
	n, rev := nrev, ""
	if len(tt) == 2 {
		n, rev = tt[0], tt[1]
	}

	h.Stop()
	defer h.Start()
	msg := fmt.Sprintf("RollingBack chart [yellow::b]%s[-::-] to release <[orangered::b]%s[-::-]>?", n, rev)
	dialog.ShowConfirmAck(h.App().App, h.App().Content.Pages, n, false, "Confirm Rollback", msg, func() {
		ctx, cancel := context.WithTimeout(context.Background(), h.App().Conn().Config().CallTimeout())
		defer cancel()
		if err := h.rollback(ctx, client.FQN(ns, n), rev); err != nil {
			h.App().Flash().Err(err)
		} else {
			h.App().Flash().Infof("Rollout restart in progress for char `%s...", n)
		}
	}, func() {})

	return nil
}

func (h *History) rollback(ctx context.Context, path, rev string) error {
	var hm dao.HelmHistory
	hm.Init(h.App().factory, h.GVR())
	if err := hm.Rollback(ctx, path, rev); err != nil {
		return err
	}
	h.Refresh()

	return nil
}
