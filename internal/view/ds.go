package view

import (
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// DaemonSet represents a daemon set custom viewer.
type DaemonSet struct {
	ResourceViewer
}

// NewDaemonSet returns a new viewer.
func NewDaemonSet(gvr client.GVR) ResourceViewer {
	d := DaemonSet{
		ResourceViewer: NewRestartExtender(
			NewLogsExtender(NewBrowser(gvr), nil),
		),
	}
	d.SetBindKeysFn(d.bindKeys)
	d.GetTable().SetEnterFn(d.showPods)
	d.GetTable().SetColorerFn(render.DaemonSet{}.ColorerFunc())

	return &d
}

func (d *DaemonSet) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyShiftD: ui.NewKeyAction("Sort Desired", d.GetTable().SortColCmd(1, true), false),
		ui.KeyShiftC: ui.NewKeyAction("Sort Current", d.GetTable().SortColCmd(2, true), false),
		ui.KeyShiftR: ui.NewKeyAction("Sort Ready", d.GetTable().SortColCmd(3, true), false),
		ui.KeyShiftU: ui.NewKeyAction("Sort UpToDate", d.GetTable().SortColCmd(4, true), false),
		ui.KeyShiftV: ui.NewKeyAction("Sort Available", d.GetTable().SortColCmd(5, true), false),
	})
}

func (d *DaemonSet) showPods(app *App, _, _, path string) {
	o, err := app.factory.Get(d.GVR(), path, true, labels.Everything())
	if err != nil {
		d.App().Flash().Err(err)
		return
	}

	var ds appsv1.DaemonSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
	if err != nil {
		d.App().Flash().Err(err)
	}

	showPodsFromSelector(app, strings.Replace(path, "/", "::", 1), ds.Spec.Selector)
}
