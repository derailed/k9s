package views

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type statefulSetView struct {
	*logResourceView
	scalableResourceView    *scalableResourceView
	restartableResourceView *restartableResourceView
}

func newStatefulSetView(title, gvr string, app *appView, list resource.List) resourceViewer {
	logResourceView := newLogResourceView(title, gvr, app, list)
	v := statefulSetView{
		logResourceView:         logResourceView,
		scalableResourceView:    newScalableResourceViewForParent(logResourceView.resourceView),
		restartableResourceView: newRestartableResourceViewForParent(logResourceView.resourceView),
	}
	v.extraActionsFn = v.extraActions
	v.enterFn = v.showPods

	return &v
}

func (v *statefulSetView) extraActions(aa ui.KeyActions) {
	v.logResourceView.extraActions(aa)
	v.scalableResourceView.extraActions(aa)
	v.restartableResourceView.extraActions(aa)
	aa[ui.KeyShiftD] = ui.NewKeyAction("Sort Desired", sortColCmd(v, 1, false), false)
	aa[ui.KeyShiftC] = ui.NewKeyAction("Sort Current", sortColCmd(v, 2, false), false)
}

func (v *statefulSetView) showPods(app *appView, ns, res, sel string) {
	ns, n := namespaced(sel)
	s := k8s.NewStatefulSet(app.Conn())
	st, err := s.Get(ns, n)
	if err != nil {
		log.Error().Err(err).Msgf("Fetching StatefulSet %s", sel)
		app.Flash().Errf("Unable to fetch statefulset %s", err)
		return
	}

	sts := st.(*v1.StatefulSet)
	l, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		log.Error().Err(err).Msgf("Converting selector for StatefulSet %s", sel)
		app.Flash().Errf("Selector failed %s", err)
		return
	}

	showPods(app, ns, l.String(), "", v.backCmd)
}
