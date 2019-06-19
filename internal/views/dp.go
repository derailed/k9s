package views

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deployView struct {
	*logResourceView
}

func newDeployView(title string, app *appView, list resource.List) resourceViewer {
	v := deployView{newLogResourceView(title, app, list)}
	v.extraActionsFn = v.extraActions
	v.enterFn = v.showPods

	return &v
}

func (v *deployView) extraActions(aa keyActions) {
	v.logResourceView.extraActions(aa)
	aa[KeyShiftD] = newKeyAction("Sort Desired", v.sortColCmd(2, false), true)
	aa[KeyShiftC] = newKeyAction("Sort Current", v.sortColCmd(3, false), true)
}

func (v *deployView) showPods(app *appView, _, res, sel string) {
	ns, n := namespaced(sel)
	d := k8s.NewDeployment(app.conn())
	dep, err := d.Get(ns, n)
	if err != nil {
		app.flash().err(err)
		return
	}

	dp := dep.(*v1.Deployment)
	l, err := metav1.LabelSelectorAsSelector(dp.Spec.Selector)
	if err != nil {
		app.flash().err(err)
		return
	}

	showPods(app, ns, l.String(), "", v.backCmd)
}
