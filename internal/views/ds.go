package views

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	extv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type daemonSetView struct {
	*logResourceView
}

func newDaemonSetView(t string, app *appView, list resource.List) resourceViewer {
	v := daemonSetView{newLogResourceView(t, app, list)}
	v.extraActionsFn = v.extraActions
	v.enterFn = v.showPods

	return &v
}

func (v *daemonSetView) extraActions(aa ui.KeyActions) {
	v.logResourceView.extraActions(aa)
	aa[ui.KeyShiftD] = ui.NewKeyAction("Sort Desired", v.sortColCmd(2, false), true)
	aa[ui.KeyShiftC] = ui.NewKeyAction("Sort Current", v.sortColCmd(3, false), true)
}

func (v *daemonSetView) showPods(app *appView, _, res, sel string) {
	ns, n := namespaced(sel)
	d := k8s.NewDaemonSet(app.Conn())
	dset, err := d.Get(ns, n)
	if err != nil {
		v.app.Flash().Err(err)
		return
	}

	ds := dset.(*extv1beta1.DaemonSet)
	l, err := metav1.LabelSelectorAsSelector(ds.Spec.Selector)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPods(app, ns, l.String(), "", v.backCmd)
}
