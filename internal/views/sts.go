package views

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type statefulSetView struct {
	*logResourceView
}

func newStatefulSetView(t string, app *appView, list resource.List) resourceViewer {
	v := statefulSetView{newLogResourceView(t, app, list)}
	v.extraActionsFn = v.extraActions
	v.enterFn = v.showPods

	return &v
}

func (v *statefulSetView) extraActions(aa keyActions) {
	v.logResourceView.extraActions(aa)
	aa[KeyShiftD] = newKeyAction("Sort Desired", v.sortColCmd(1, false), true)
	aa[KeyShiftC] = newKeyAction("Sort Current", v.sortColCmd(2, false), true)
}

func (v *statefulSetView) showPods(app *appView, ns, res, sel string) {
	ns, n := namespaced(sel)
	s := k8s.NewStatefulSet(app.conn())
	st, err := s.Get(ns, n)
	if err != nil {
		log.Error().Err(err).Msgf("Fetching StatefulSet %s", sel)
		app.flash().errf("Unable to fetch statefulset %s", err)
		return
	}

	sts := st.(*v1.StatefulSet)
	l, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		log.Error().Err(err).Msgf("Converting selector for StatefulSet %s", sel)
		app.flash().errf("Selector failed %s", err)
		return
	}

	showPods(app, ns, l.String(), "", v.backCmd)
}
