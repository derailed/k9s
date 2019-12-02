package view

import (
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const scaleDialogKey = "scale"

// Deploy represents a deployment view.
type Deploy struct {
	ResourceViewer
}

// NewDeploy returns a new deployment view.
func NewDeploy(gvr dao.GVR) ResourceViewer {
	d := Deploy{
		ResourceViewer: NewRestartExtender(
			NewScaleExtender(
				NewLogsExtender(
					NewGeneric(gvr),
					func() string { return "" },
				),
			),
		),
	}
	d.BindKeys()
	d.GetTable().SetEnterFn(d.showPods)
	d.GetTable().SetColorerFn(render.Deployment{}.ColorerFunc())

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
	o, err := app.factory.Get(ns, d.GVR(), n, labels.Everything())
	if err != nil {
		app.Flash().Err(err)
		return
	}

	var dp appsv1.Deployment
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &dp)
	if err != nil {
		app.Flash().Err(err)
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
