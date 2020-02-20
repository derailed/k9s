package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const scaleDialogKey = "scale"

// Deploy represents a deployment view.
type Deploy struct {
	ResourceViewer
}

// NewDeploy returns a new deployment view.
func NewDeploy(gvr client.GVR) ResourceViewer {
	d := Deploy{
		ResourceViewer: NewPortForwardExtender(
			NewRestartExtender(
				NewScaleExtender(
					NewLogsExtender(
						NewBrowser(gvr),
						nil,
					),
				),
			),
		),
	}
	d.SetBindKeysFn(d.bindKeys)
	d.GetTable().SetEnterFn(d.showPods)
	d.GetTable().SetColorerFn(render.Deployment{}.ColorerFunc())

	return &d
}

func (d *Deploy) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyShiftR: ui.NewKeyAction("Sort Ready", d.GetTable().SortColCmd(1, true), false),
		ui.KeyShiftU: ui.NewKeyAction("Sort UpToDate", d.GetTable().SortColCmd(2, true), false),
		ui.KeyShiftL: ui.NewKeyAction("Sort Available", d.GetTable().SortColCmd(3, true), false),
	})
}

func (d *Deploy) showPods(app *App, model ui.Tabular, gvr, path string) {
	var res dao.Deployment
	res.Init(app.factory, client.NewGVR(d.GVR()))

	dp, err := res.GetInstance(path)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPodsFromSelector(app, path, dp.Spec.Selector)
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
