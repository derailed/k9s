package view

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type DaemonSet struct {
	ResourceViewer
}

func NewDaemonSet(title, gvr string, list resource.List) ResourceViewer {
	d := DaemonSet{
		ResourceViewer: NewRestartExtender(
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

func (d *DaemonSet) BindKeys() {
	d.Actions().Add(ui.KeyActions{
		ui.KeyShiftD: ui.NewKeyAction("Sort Desired", d.GetTable().SortColCmd(1, true), false),
		ui.KeyShiftC: ui.NewKeyAction("Sort Current", d.GetTable().SortColCmd(2, true), false),
	})
}

func (d *DaemonSet) showPods(app *App, _, res, sel string) {
	ns, n := namespaced(sel)
	dset, err := k8s.NewDaemonSet(app.Conn()).Get(ns, n)
	if err != nil {
		d.App().Flash().Err(err)
		return
	}

	ds, ok := dset.(*appsv1.DaemonSet)
	if !ok {
		log.Fatal().Msg("Expecting a valid ds")
	}
	l, err := metav1.LabelSelectorAsSelector(ds.Spec.Selector)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPods(app, ns, l.String(), "")
}
