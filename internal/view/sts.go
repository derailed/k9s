package view

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatefulSet represents a statefulset viewer.
type StatefulSet struct {
	*LogResource
	scalableResource    *ScalableResource
	restartableResource *RestartableResource
}

// NewStatefulSet returns a new viewer.
func NewStatefulSet(title, gvr string, list resource.List) ResourceViewer {
	l := NewLogResource(title, gvr, list)
	s := StatefulSet{
		LogResource:         l,
		scalableResource:    newScalableResourceForParent(l.Resource),
		restartableResource: newRestartableResourceForParent(l.Resource),
	}
	s.extraActionsFn = s.extraActions
	s.enterFn = s.showPods

	return &s
}

func (s *StatefulSet) extraActions(aa ui.KeyActions) {
	s.LogResource.extraActions(aa)
	s.scalableResource.extraActions(aa)
	s.restartableResource.extraActions(aa)
	aa[ui.KeyShiftD] = ui.NewKeyAction("Sort Desired", s.sortColCmd(1), false)
	aa[ui.KeyShiftC] = ui.NewKeyAction("Sort Current", s.sortColCmd(2), false)
}

func (s *StatefulSet) showPods(app *App, _, res, sel string) {
	ns, n := namespaced(sel)
	st, err := k8s.NewStatefulSet(app.Conn()).Get(ns, n)
	if err != nil {
		log.Error().Err(err).Msgf("Fetching StatefulSet %s", sel)
		app.Flash().Errf("Unable to fetch statefulset %s", err)
		return
	}

	sts, ok := st.(*v1.StatefulSet)
	if !ok {
		log.Fatal().Msg("Expecting a valid sts")
	}
	l, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		log.Error().Err(err).Msgf("Converting selector for StatefulSet %s", sel)
		app.Flash().Errf("Selector failed %s", err)
		return
	}

	showPods(app, ns, l.String(), "")
}
