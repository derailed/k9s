package views

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type deployView struct {
	*logResourceView
	scalableResourceView    *scalableResourceView
	restartableResourceView *restartableResourceView
}

const scaleDialogKey = "scale"

func newDeployView(title, gvr string, app *appView, list resource.List) resourceViewer {
	logResourceView := newLogResourceView(title, gvr, app, list)
	v := deployView{
		logResourceView:         logResourceView,
		scalableResourceView:    newScalableResourceViewForParent(logResourceView.resourceView),
		restartableResourceView: newRestartableResourceViewForParent(logResourceView.resourceView),
	}
	v.extraActionsFn = v.extraActions
	v.enterFn = v.showPods

	return &v
}

func (v *deployView) extraActions(aa ui.KeyActions) {
	v.logResourceView.extraActions(aa)
	v.scalableResourceView.extraActions(aa)
	v.restartableResourceView.extraActions(aa)
	aa[ui.KeyShiftD] = ui.NewKeyAction("Sort Desired", v.sortColCmd(1, false), false)
	aa[ui.KeyShiftC] = ui.NewKeyAction("Sort Current", v.sortColCmd(2, false), false)
}

func (v *deployView) showPods(app *appView, _, res, sel string) {
	ns, n := namespaced(sel)
	d := k8s.NewDeployment(app.Conn())
	dep, err := d.Get(ns, n)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	dp := dep.(*v1.Deployment)
	l, err := metav1.LabelSelectorAsSelector(dp.Spec.Selector)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPods(app, ns, l.String(), "", v.backCmd)
}
