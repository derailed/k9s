package view

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const scaleDialogKey = "scale"

// Deploy represents a deployment view.
type Deploy struct {
	ResourceViewer
}

// NewDeploy returns a new deployment view.
func NewDeploy(title, gvr string, list resource.List) ResourceViewer {
	d := Deploy{
		ResourceViewer: NewRestartExtender(
			NewScaleExtender(
				NewLogsExtender(
					NewResource(title, gvr, list),
					func() string { return "" },
				),
			),
		),
	}
	d.BindKeys()
	d.GetTable().SetEnterFn(d.showPods)

	return &d
}

func (d *Deploy) BindKeys() {
	d.Actions().Add(ui.KeyActions{
		ui.KeyShiftD: ui.NewKeyAction("Sort Desired", d.GetTable().SortColCmd(1, true), false),
		ui.KeyShiftC: ui.NewKeyAction("Sort Current", d.GetTable().SortColCmd(2, true), false),
	})
}

func (d *Deploy) showPods(app *App, _, res, sel string) {
	ns, n := namespaced(sel)
	dep, err := k8s.NewDeployment(app.Conn()).Get(ns, n)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	dp, ok := dep.(*v1.Deployment)
	if !ok {
		log.Fatal().Msg("Expecting valid deployment")
	}
	showPodsFromSelector(app, ns, dp.Spec.Selector)
}

// Helpers...

func showPodsFromSelector(app *App, ns string, sel *metav1.LabelSelector) {
	l, err := metav1.LabelSelectorAsSelector(sel)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPods(app, ns, l.String(), "")
}
