package views

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deployView struct {
	*resourceView
}

func newDeployView(t string, app *appView, list resource.List) resourceViewer {
	v := deployView{newResourceView(t, app, list).(*resourceView)}
	v.extraActionsFn = v.extraActions

	return &v
}

func (v *deployView) extraActions(aa keyActions) {
	aa[KeyShiftD] = newKeyAction("Sort Desired", v.sortColCmd(2, false), true)
	aa[KeyShiftC] = newKeyAction("Sort Current", v.sortColCmd(3, false), true)
	aa[tcell.KeyEnter] = newKeyAction("View Pods", v.showPodsCmd, true)
}

func (v *deployView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.getTV()
		t.sortCol.index, t.sortCol.asc = t.nameColIndex()+col, asc
		t.refresh()

		return nil
	}
}

func (v *deployView) showPodsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	ns, n := namespaced(v.selectedItem)
	d := k8s.NewDeployment(v.app.conn())
	dep, err := d.Get(ns, n)
	if err != nil {
		log.Error().Err(err).Msgf("Fetching Deployment %s", v.selectedItem)
		v.app.flash().err(err)
		return evt
	}
	dp := dep.(*v1.Deployment)

	sel, err := metav1.LabelSelectorAsSelector(dp.Spec.Selector)
	if err != nil {
		log.Error().Err(err).Msgf("Converting selector for Deployment %s", v.selectedItem)
		v.app.flash().err(err)
		return evt
	}
	showPods(v.app, ns, "Deployment", v.selectedItem, sel.String(), "", v.backCmd)

	return nil
}

func (v *deployView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	// Reset namespace to what it was
	v.app.config.SetActiveNamespace(v.list.GetNamespace())
	v.app.inject(v)

	return nil
}
