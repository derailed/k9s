package view

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

// ReplicationController represents a deployment view.
type ReplicationController struct {
	ResourceViewer
}

// NewReplicationController returns a new deployment view.
func NewReplicationController(title, gvr string, list resource.List) ResourceViewer {
	d := ReplicationController{
		ResourceViewer: NewScaleExtender(
			NewLogsExtender(
				NewResource(title, gvr, list),
				func() string { return "" },
			),
		),
	}
	d.BindKeys()
	d.GetTable().SetEnterFn(d.showPods)

	return &d
}

func (d *ReplicationController) BindKeys() {
	d.Actions().Add(ui.KeyActions{
		ui.KeyShiftD: ui.NewKeyAction("Sort Desired", d.GetTable().SortColCmd(1, true), false),
		ui.KeyShiftC: ui.NewKeyAction("Sort Current", d.GetTable().SortColCmd(2, true), false),
	})
}

func (d *ReplicationController) showPods(app *App, _, res, sel string) {
	ns, n := namespaced(sel)
	nrc, err := k8s.NewReplicationController(app.Conn()).Get(ns, n)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	rc, ok := nrc.(*v1.ReplicationController)
	if !ok {
		log.Fatal().Msg("Expecting valid replication controller")
	}
	showPodsFromLabels(app, ns, rc.Spec.Selector)
}
