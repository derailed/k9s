// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"errors"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	appsv1 "k8s.io/api/apps/v1"
)

// DaemonSet represents a daemon set custom viewer.
type DaemonSet struct {
	ResourceViewer
}

// NewDaemonSet returns a new viewer.
func NewDaemonSet(gvr client.GVR) ResourceViewer {
	var d DaemonSet
	d.ResourceViewer = NewPortForwardExtender(
		NewVulnerabilityExtender(
			NewRestartExtender(
				NewImageExtender(
					NewOwnerExtender(
						NewLogsExtender(NewBrowser(gvr), d.logOptions),
					),
				),
			),
		),
	)
	d.AddBindKeysFn(d.bindKeys)
	d.GetTable().SetEnterFn(d.showPods)

	return &d
}

func (d *DaemonSet) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyShiftD: ui.NewKeyAction("Sort Desired", d.GetTable().SortColCmd("DESIRED", true), false),
		ui.KeyShiftC: ui.NewKeyAction("Sort Current", d.GetTable().SortColCmd("CURRENT", true), false),
		ui.KeyShiftR: ui.NewKeyAction("Sort Ready", d.GetTable().SortColCmd(readyCol, true), false),
		ui.KeyShiftU: ui.NewKeyAction("Sort UpToDate", d.GetTable().SortColCmd(uptodateCol, true), false),
		ui.KeyShiftL: ui.NewKeyAction("Sort Available", d.GetTable().SortColCmd(availCol, true), false),
	})
}

func (d *DaemonSet) showPods(app *App, model ui.Tabular, _ client.GVR, path string) {
	var res dao.DaemonSet
	res.Init(app.factory, d.GVR())

	ds, err := res.GetInstance(path)
	if err != nil {
		d.App().Flash().Err(err)
		return
	}

	showPodsFromSelector(app, path, ds.Spec.Selector)
}

func (d *DaemonSet) logOptions(prev bool) (*dao.LogOptions, error) {
	path := d.GetTable().GetSelectedItem()
	if path == "" {
		return nil, errors.New("you must provide a selection")
	}
	ds, err := d.getInstance(path)
	if err != nil {
		return nil, err
	}

	return podLogOptions(d.App(), path, prev, ds.ObjectMeta, ds.Spec.Template.Spec), nil
}

func (d *DaemonSet) getInstance(fqn string) (*appsv1.DaemonSet, error) {
	var ds dao.DaemonSet
	ds.Init(d.App().factory, client.NewGVR("apps/v1/daemonsets"))

	return ds.GetInstance(fqn)
}
