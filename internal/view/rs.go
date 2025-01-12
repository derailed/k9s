// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
)

// ReplicaSet presents a replicaset viewer.
type ReplicaSet struct {
	ResourceViewer
}

// NewReplicaSet returns a new viewer.
func NewReplicaSet(gvr client.GVR) ResourceViewer {
	r := ReplicaSet{
		ResourceViewer: NewOwnerExtender(
			NewVulnerabilityExtender(
				NewBrowser(gvr),
			),
		),
	}
	r.AddBindKeysFn(r.bindKeys)
	r.GetTable().SetEnterFn(r.showPods)

	return &r
}

func (r *ReplicaSet) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyShiftD:   ui.NewKeyAction("Sort Desired", r.GetTable().SortColCmd("DESIRED", true), false),
		ui.KeyShiftC:   ui.NewKeyAction("Sort Current", r.GetTable().SortColCmd("CURRENT", true), false),
		ui.KeyShiftR:   ui.NewKeyAction("Sort Ready", r.GetTable().SortColCmd(readyCol, true), false),
		tcell.KeyCtrlL: ui.NewKeyAction("Rollback", r.rollbackCmd, true),
	})
}

func (r *ReplicaSet) showPods(app *App, _ ui.Tabular, _ client.GVR, path string) {
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

	msg := fmt.Sprintf("Rollback %s %s?", r.GVR(), path)

	dialog.ShowConfirm(r.App().Styles.Dialog(), r.App().Content.Pages, "Rollback", msg, func() {
		r.App().Flash().Infof("Rolling back %s %s", r.GVR(), path)
		var drs dao.ReplicaSet
		drs.Init(r.App().factory, r.GVR())
		if err := drs.Rollback(path); err != nil {
			r.App().Flash().Err(err)
		} else {
			r.App().Flash().Infof("%s successfully rolled back", path)
		}
		r.Refresh()
	}, func() {})

	return nil
}
