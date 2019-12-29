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

// StatefulSet represents a statefulset viewer.
type StatefulSet struct {
	ResourceViewer
}

// NewStatefulSet returns a new viewer.
func NewStatefulSet(gvr client.GVR) ResourceViewer {
	s := StatefulSet{
		ResourceViewer: NewRestartExtender(
			NewScaleExtender(
				NewLogsExtender(NewBrowser(gvr), nil),
			),
		),
	}
	s.SetBindKeysFn(s.bindKeys)
	s.GetTable().SetEnterFn(s.showPods)
	s.GetTable().SetColorerFn(render.StatefulSet{}.ColorerFunc())

	return &s
}

func (s *StatefulSet) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyShiftD: ui.NewKeyAction("Sort Desired", s.GetTable().SortColCmd(1, true), false),
		ui.KeyShiftC: ui.NewKeyAction("Sort Current", s.GetTable().SortColCmd(2, true), false),
	})
}

func (s *StatefulSet) showPods(app *App, _, gvr, path string) {
	o, err := app.factory.Get(s.GVR(), path, labels.Everything())
	if err != nil {
		app.Flash().Err(err)
		return
	}

	var sts appsv1.StatefulSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &sts)
	if err != nil {
		app.Flash().Err(err)
	}

	showPodsFromSelector(app, strings.Replace(path, "/", "::", 1), sts.Spec.Selector)
}
