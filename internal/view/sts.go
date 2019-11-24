package view

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/apps/v1"
)

// StatefulSet represents a statefulset viewer.
type StatefulSet struct {
	ResourceViewer
}

// NewStatefulSet returns a new viewer.
func NewStatefulSet(title, gvr string, list resource.List) ResourceViewer {
	s := StatefulSet{
		ResourceViewer: NewRestartExtender(
			NewScaleExtender(
				NewLogsExtender(
					NewResource(title, gvr, list),
					func() string { return "" },
				),
			),
		),
	}
	s.BindKeys()
	s.GetTable().SetEnterFn(s.showPods)

	return &s
}

func (d *StatefulSet) BindKeys() {
	d.Actions().Add(ui.KeyActions{
		ui.KeyShiftD: ui.NewKeyAction("Sort Desired", d.GetTable().SortColCmd(1, true), false),
		ui.KeyShiftC: ui.NewKeyAction("Sort Current", d.GetTable().SortColCmd(2, true), false),
	})
}

func (s *StatefulSet) showPods(app *App, _, res, path string) {
	ns, n := namespaced(path)
	st, err := k8s.NewStatefulSet(app.Conn()).Get(ns, n)
	if err != nil {
		log.Error().Err(err).Msgf("Fetching StatefulSet %s", path)
		app.Flash().Errf("Unable to fetch statefulset %s", err)
		return
	}

	sts, ok := st.(*v1.StatefulSet)
	if !ok {
		log.Fatal().Msg("Expecting a valid sts")
	}
	showPodsFromSelector(app, ns, sts.Spec.Selector)
}
