package view

import (
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
		ResourceViewer: NewPortForwardExtender(
			NewRestartExtender(
				NewScaleExtender(
					NewSetImageExtender(
						NewLogsExtender(NewBrowser(gvr), nil),
					),
				),
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
		ui.KeyShiftR: ui.NewKeyAction("Sort Ready", s.GetTable().SortColCmd(readyCol, true), false),
	})
}

func (s *StatefulSet) showPods(app *App, _ ui.Tabular, _, path string) {
	sts, err := s.sts(path)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPodsFromSelector(app, path, sts.Spec.Selector)
}

func (s *StatefulSet) sts(path string) (*appsv1.StatefulSet, error) {
	o, err := s.App().factory.Get(s.GVR().String(), path, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var sts appsv1.StatefulSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &sts)
	if err != nil {
		return nil, err
	}

	return &sts, nil
}
