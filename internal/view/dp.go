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

	sts, err := d.dp(path)
	if err != nil {
		return nil, err
	}

	cc := sts.Spec.Template.Spec.Containers
	var (
		co, dco string
		allCos  bool
	)
	if c, ok := dao.GetDefaultContainer(sts.Spec.Template.ObjectMeta, sts.Spec.Template.Spec); ok {
		co, dco = c, c
	} else if len(cc) == 1 {
		co = cc[0].Name
	} else {
		dco, allCos = cc[0].Name, true
	}

	cfg := d.App().Config.K9s.Logger
	opts := dao.LogOptions{
		Path:            path,
		Container:       co,
		Lines:           int64(cfg.TailCount),
		SinceSeconds:    cfg.SinceSeconds,
		SingleContainer: len(cc) == 1,
		AllContainers:   allCos,
		ShowTimestamp:   cfg.ShowTime,
		Previous:        prev,
	}
	if co == "" {
		opts.AllContainers = true
	}
	opts.DefaultContainer = dco

	return &opts, nil
}

func (d *Deploy) showPods(app *App, model ui.Tabular, gvr client.GVR, fqn string) {
	var ddp dao.Deployment
	ddp.Init(d.App().factory, d.GVR())

	dp, err := ddp.GetInstance(fqn)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPodsFromSelector(app, fqn, dp.Spec.Selector)
}

func (d *Deploy) dp(fqn string) (*appsv1.Deployment, error) {
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
