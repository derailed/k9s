package view

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

// ReplicaSet presents a replicaset viewer.
type ReplicaSet struct {
	ResourceViewer
}

// NewReplicaSet returns a new viewer.
func NewReplicaSet(gvr client.GVR) ResourceViewer {
	r := ReplicaSet{
		ResourceViewer: NewBrowser(gvr),
	}
	r.SetBindKeysFn(r.bindKeys)
	r.GetTable().SetEnterFn(r.showPods)
	r.GetTable().SetColorerFn(render.ReplicaSet{}.ColorerFunc())

	return &r
}

func (r *ReplicaSet) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyShiftD:   ui.NewKeyAction("Sort Desired", r.GetTable().SortColCmd("DESIRED", true), false),
		ui.KeyShiftC:   ui.NewKeyAction("Sort Current", r.GetTable().SortColCmd("CURRENT", true), false),
		ui.KeyShiftR:   ui.NewKeyAction("Sort Ready", r.GetTable().SortColCmd(readyCol, true), false),
		tcell.KeyCtrlL: ui.NewKeyAction("Rollback", r.rollbackCmd, true),
	})
}

func (r *ReplicaSet) showPods(app *App, model ui.Tabular, gvr, path string) {
	var drs dao.ReplicaSet
	rs, err := drs.Load(app.factory, path)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPodsFromSelector(app, path, rs.Spec.Selector)
}

func (r *ReplicaSet) rollbackCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := r.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	r.showModal(fmt.Sprintf("Rollback %s %s?", r.GVR(), path), func(_ int, button string) {
		defer r.dismissModal()

		if button != "OK" {
			return
		}
		r.App().Flash().Infof("Rolling back %s %s", r.GVR(), path)
		var drs dao.ReplicaSet
		drs.Init(r.App().factory, r.GVR())
		if err := drs.Rollback(path); err != nil {
			r.App().Flash().Err(err)
		} else {
			r.App().Flash().Infof("%s successfully rolled back", path)
		}
		r.Refresh()
	})

	return nil
}

func (r *ReplicaSet) dismissModal() {
	r.App().Content.RemovePage("confirm")
}

func (r *ReplicaSet) showModal(msg string, done func(int, string)) {
	confirm := tview.NewModal().
		AddButtons([]string{"Cancel", "OK"}).
		SetTextColor(tcell.ColorFuchsia).
		SetText(msg).
		SetDoneFunc(done)
	r.App().Content.AddPage("confirm", confirm, false, false)
	r.App().Content.ShowPage("confirm")
}
