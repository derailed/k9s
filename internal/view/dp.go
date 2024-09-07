// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"errors"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const scaleDialogKey = "scale"

// Deploy represents a deployment view.
type Deploy struct {
	ResourceViewer
}

// NewDeploy returns a new deployment view.
func NewDeploy(gvr client.GVR) ResourceViewer {
	var d Deploy
	d.ResourceViewer = NewPortForwardExtender(
		NewVulnerabilityExtender(
			NewRestartExtender(
				NewScaleExtender(
					NewImageExtender(
						NewOwnerExtender(
							NewLogsExtender(NewBrowser(gvr), d.logOptions),
						),
					),
				),
			),
		),
	)
	d.AddBindKeysFn(d.bindKeys)
	d.GetTable().SetEnterFn(d.showPods)

	return &d
}

func (d *Deploy) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyShiftR: ui.NewKeyAction("Sort Ready", d.GetTable().SortColCmd(readyCol, true), false),
		ui.KeyShiftU: ui.NewKeyAction("Sort UpToDate", d.GetTable().SortColCmd(uptodateCol, true), false),
		ui.KeyShiftL: ui.NewKeyAction("Sort Available", d.GetTable().SortColCmd(availCol, true), false),
	})
}

func (d *Deploy) logOptions(prev bool) (*dao.LogOptions, error) {
	path := d.GetTable().GetSelectedItem()
	if path == "" {
		return nil, errors.New("you must provide a selection")
	}
	dp, err := d.getInstance(path)
	if err != nil {
		return nil, err
	}

	return podLogOptions(d.App(), path, prev, dp.ObjectMeta, dp.Spec.Template.Spec), nil
}

func (d *Deploy) showPods(app *App, model ui.Tabular, gvr client.GVR, fqn string) {
	dp, err := d.getInstance(fqn)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPodsFromSelector(app, fqn, dp.Spec.Selector)
}

func (d *Deploy) getInstance(fqn string) (*appsv1.Deployment, error) {
	var dp dao.Deployment
	dp.Init(d.App().factory, d.GVR())

	return dp.GetInstance(fqn)
}

// ----------------------------------------------------------------------------
// Helpers...

func showPodsFromSelector(app *App, path string, sel *metav1.LabelSelector) {
	l, err := metav1.LabelSelectorAsSelector(sel)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPods(app, path, l.String(), "")
}
