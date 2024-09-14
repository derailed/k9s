// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Workload presents a workload viewer.
type Workload struct {
	ResourceViewer
}

// NewWorkload returns a new viewer.
func NewWorkload(gvr client.GVR) ResourceViewer {
	w := Workload{
		ResourceViewer: NewBrowser(gvr),
	}
	w.GetTable().SetEnterFn(w.showRes)
	w.AddBindKeysFn(w.bindKeys)
	w.GetTable().SetSortCol("KIND", true)

	return &w
}

func (w *Workload) bindDangerousKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyE: ui.NewKeyActionWithOpts("Edit", w.editCmd,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			}),
		tcell.KeyCtrlD: ui.NewKeyActionWithOpts("Delete", w.deleteCmd,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			}),
	})
}

func (w *Workload) bindKeys(aa *ui.KeyActions) {
	if !w.App().Config.K9s.IsReadOnly() {
		w.bindDangerousKeys(aa)
	}

	aa.Bulk(ui.KeyMap{
		ui.KeyShiftK: ui.NewKeyAction("Sort Kind", w.GetTable().SortColCmd("KIND", true), false),
		ui.KeyShiftS: ui.NewKeyAction("Sort Status", w.GetTable().SortColCmd(statusCol, true), false),
		ui.KeyShiftR: ui.NewKeyAction("Sort Ready", w.GetTable().SortColCmd("READY", true), false),
		ui.KeyShiftA: ui.NewKeyAction("Sort Age", w.GetTable().SortColCmd(ageCol, true), false),
		ui.KeyY:      ui.NewKeyAction(yamlAction, w.yamlCmd, true),
		ui.KeyD:      ui.NewKeyAction("Describe", w.describeCmd, true),
	})
}

func parsePath(path string) (client.GVR, string, bool) {
	tt := strings.Split(path, "|")
	if len(tt) != 3 {
		log.Error().Msgf("unable to parse path: %q", path)
		return client.NewGVR(""), client.FQN("", ""), false
	}

	return client.NewGVR(tt[0]), client.FQN(tt[1], tt[2]), true
}

func (w *Workload) showRes(app *App, _ ui.Tabular, _ client.GVR, path string) {
	gvr, fqn, ok := parsePath(path)
	if !ok {
		app.Flash().Err(fmt.Errorf("unable to parse path: %q", path))
		return
	}
	app.gotoResource(gvr.R(), fqn, false)
}

func (w *Workload) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	selections := w.GetTable().GetSelectedItems()
	if len(selections) == 0 {
		return evt
	}

	w.Stop()
	defer w.Start()
	{
		msg := fmt.Sprintf("Delete %s %s?", w.GVR().R(), selections[0])
		if len(selections) > 1 {
			msg = fmt.Sprintf("Delete %d marked %s?", len(selections), w.GVR())
		}
		w.resourceDelete(selections, msg)
	}

	return nil
}

func (w *Workload) defaultContext(gvr client.GVR, fqn string) context.Context {
	ctx := context.WithValue(context.Background(), internal.KeyFactory, w.App().factory)
	ctx = context.WithValue(ctx, internal.KeyGVR, gvr)
	if fqn != "" {
		ctx = context.WithValue(ctx, internal.KeyPath, fqn)
	}
	if internal.IsLabelSelector(w.GetTable().CmdBuff().GetText()) {
		ctx = context.WithValue(ctx, internal.KeyLabels, ui.TrimLabelSelector(w.GetTable().CmdBuff().GetText()))
	}
	ctx = context.WithValue(ctx, internal.KeyNamespace, client.CleanseNamespace(w.App().Config.ActiveNamespace()))
	ctx = context.WithValue(ctx, internal.KeyWithMetrics, w.App().factory.Client().HasMetrics())

	return ctx
}

func (w *Workload) resourceDelete(selections []string, msg string) {
	okFn := func(propagation *metav1.DeletionPropagation, force bool) {
		w.GetTable().ShowDeleted()
		if len(selections) > 1 {
			w.App().Flash().Infof("Delete %d marked %s", len(selections), w.GVR())
		} else {
			w.App().Flash().Infof("Delete resource %s %s", w.GVR(), selections[0])
		}
		for _, sel := range selections {
			gvr, fqn, ok := parsePath(sel)
			if !ok {
				w.App().Flash().Err(fmt.Errorf("unable to parse path: %q", sel))
				return
			}

			grace := dao.DefaultGrace
			if force {
				grace = dao.ForceGrace
			}
			if err := w.GetTable().GetModel().Delete(w.defaultContext(gvr, fqn), fqn, propagation, grace); err != nil {
				w.App().Flash().Errf("Delete failed with `%s", err)
			} else {
				w.App().factory.DeleteForwarder(sel)
			}
			w.GetTable().DeleteMark(sel)
		}
		w.GetTable().Start()
	}
	dialog.ShowDelete(w.App().Styles.Dialog(), w.App().Content.Pages, msg, okFn, func() {})
}

func (w *Workload) describeCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := w.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}
	gvr, fqn, ok := parsePath(path)
	if !ok {
		w.App().Flash().Err(fmt.Errorf("unable to parse path: %q", path))
		return evt
	}

	describeResource(w.App(), nil, gvr, fqn)

	return nil
}

func (w *Workload) editCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := w.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}
	gvr, fqn, ok := parsePath(path)
	if !ok {
		w.App().Flash().Err(fmt.Errorf("unable to parse path: %q", path))
		return evt
	}

	w.Stop()
	defer w.Start()
	if err := editRes(w.App(), gvr, fqn); err != nil {
		w.App().Flash().Err(err)
	}

	return nil
}

func (w *Workload) yamlCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := w.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}
	gvr, fqn, ok := parsePath(path)
	if !ok {
		w.App().Flash().Err(fmt.Errorf("unable to parse path: %q", path))
		return evt
	}

	v := NewLiveView(w.App(), yamlAction, model.NewYAML(gvr, fqn))
	if err := v.app.inject(v, false); err != nil {
		v.app.Flash().Err(err)
	}

	return nil
}
