package views

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type jobView struct {
	*logResourceView
}

func newJobView(title, gvr string, app *appView, list resource.List) resourceViewer {
	v := jobView{newLogResourceView(title, gvr, app, list)}
	v.extraActionsFn = v.extraActions
	v.enterFn = v.showPods

	return &v
}

func (v *jobView) extraActions(aa ui.KeyActions) {
	v.logResourceView.extraActions(aa)
}

func (v *jobView) showPods(app *appView, ns, res, sel string) {
	ns, n := namespaced(sel)
	j := k8s.NewJob(app.Conn())
	job, err := j.Get(ns, n)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	jo := job.(*batchv1.Job)
	l, err := metav1.LabelSelectorAsSelector(jo.Spec.Selector)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPods(app, ns, l.String(), "", v.backCmd)
}
