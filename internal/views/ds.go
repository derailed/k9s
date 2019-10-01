package views

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type daemonSetView struct {
	*logResourceView
}

func newDaemonSetView(title, gvr string, app *appView, list resource.List) resourceViewer {
	v := daemonSetView{newLogResourceView(title, gvr, app, list)}
	v.extraActionsFn = v.extraActions
	v.enterFn = v.showPods

	return &v
}

func (v *daemonSetView) extraActions(aa ui.KeyActions) {
	v.logResourceView.extraActions(aa)
	aa[ui.KeyShiftD] = ui.NewKeyAction("Sort Desired", v.sortColCmd(2, false), false)
	aa[ui.KeyShiftC] = ui.NewKeyAction("Sort Current", v.sortColCmd(3, false), false)
}

func (v *daemonSetView) showPods(app *appView, _, res, sel string) {
	ns, n := namespaced(sel)
	d := k8s.NewDaemonSet(app.Conn())
	dset, err := d.Get(ns, n)
	if err != nil {
		v.app.Flash().Err(err)
		return
	}

	ds := dset.(*appsv1.DaemonSet)
	l, err := metav1.LabelSelectorAsSelector(ds.Spec.Selector)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPods(app, ns, l.String(), "", v.backCmd)
}
