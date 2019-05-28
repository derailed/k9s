package views

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type statefulSetView struct {
	*resourceView
}

func newStatefulSetView(t string, app *appView, list resource.List) resourceViewer {
	v := statefulSetView{newResourceView(t, app, list).(*resourceView)}
	{
		v.extraActionsFn = v.extraActions
		v.switchPage("sts")
	}

	return &v
}

func (v *statefulSetView) extraActions(aa keyActions) {
	aa[KeyShiftD] = newKeyAction("Sort Desired", v.sortColCmd(1, false), true)
	aa[KeyShiftC] = newKeyAction("Sort Current", v.sortColCmd(2, false), true)
	aa[tcell.KeyEnter] = newKeyAction("View Pods", v.showPodsCmd, true)
}

func (v *statefulSetView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.getTV()
		t.sortCol.index, t.sortCol.asc = t.nameColIndex()+col, asc
		t.refresh()

		return nil
	}
}

func (v *statefulSetView) showPodsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	ns, n := namespaced(v.selectedItem)
	d := k8s.NewStatefulSet(v.app.conn())
	s, err := d.Get(ns, n)
	if err != nil {
		log.Error().Err(err).Msgf("Fetching StatefulSet %s", v.selectedItem)
		v.app.flash().errf("Unable to fetch statefulset %s", err)
		return evt
	}
	sts := s.(*v1.StatefulSet)

	sel, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		log.Error().Err(err).Msgf("Converting selector for StatefulSet %s", v.selectedItem)
		v.app.flash().errf("Selector failed %s", err)
		return evt
	}
	showPods(v.app, "", "StatefulSet", v.selectedItem, sel.String(), "", v.backCmd)

	return nil
}

func (v *statefulSetView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.app.inject(v)

	return nil
}
