package views

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type daemonSetView struct {
	*resourceView
}

func newDaemonSetView(t string, app *appView, list resource.List) resourceViewer {
	v := daemonSetView{newResourceView(t, app, list).(*resourceView)}
	v.extraActionsFn = v.extraActions
	v.enterFn = v.showPods

	return &v
}

func (v *daemonSetView) extraActions(aa keyActions) {
	aa[KeyShiftD] = newKeyAction("Sort Desired", v.sortColCmd(2, false), true)
	aa[KeyShiftC] = newKeyAction("Sort Current", v.sortColCmd(3, false), true)
}

func (v *daemonSetView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.getTV()
		t.sortCol.index, t.sortCol.asc = t.nameColIndex()+col, asc
		t.refresh()

		return nil
	}
}

func (v *daemonSetView) showPods(app *appView, _, res, sel string) {
	ns, n := namespaced(sel)
	d := k8s.NewDaemonSet(app.conn())
	dset, err := d.Get(ns, n)
	if err != nil {
		log.Error().Err(err).Msgf("Fetching DeaemonSet %s", sel)
		v.app.flash().err(err)
		return
	}

	ds := dset.(*extv1beta1.DaemonSet)
	l, err := metav1.LabelSelectorAsSelector(ds.Spec.Selector)
	if err != nil {
		log.Error().Err(err).Msgf("Converting selector for DaemonSet %s", sel)
		app.flash().err(err)
		return
	}

	showPods(app, ns, "DaemonSet", sel, l.String(), "", v.backCmd)
}

func (v *daemonSetView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	// Reset namespace to what it was
	v.app.config.SetActiveNamespace(v.list.GetNamespace())
	v.app.inject(v)

	return nil
}
