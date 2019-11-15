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
	*LogResource

	scalableResource    *ScalableResource
	restartableResource *RestartableResource
}

// NewDeploy returns a new deployment view.
func NewDeploy(title, gvr string, list resource.List) ResourceViewer {
	l := NewLogResource(title, gvr, list)
	d := Deploy{
		LogResource:         l,
		scalableResource:    newScalableResourceForParent(l.Resource),
		restartableResource: newRestartableResourceForParent(l.Resource),
	}
	d.extraActionsFn = d.extraActions
	d.enterFn = d.showPods

	return &d
}

func (d *Deploy) extraActions(aa ui.KeyActions) {
	d.LogResource.extraActions(aa)
	d.scalableResource.extraActions(aa)
	d.restartableResource.extraActions(aa)
	aa[ui.KeyShiftD] = ui.NewKeyAction("Sort Desired", d.sortColCmd(1), false)
	aa[ui.KeyShiftC] = ui.NewKeyAction("Sort Current", d.sortColCmd(2), false)
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
	l, err := metav1.LabelSelectorAsSelector(dp.Spec.Selector)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPods(app, ns, l.String(), "")
}
