package view

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Job represents a job viewer.
type Job struct {
	*LogResource
}

// NewJob returns a new viewer.
func NewJob(title, gvr string, list resource.List) ResourceViewer {
	j := Job{NewLogResource(title, gvr, list)}
	j.extraActionsFn = j.extraActions
	j.enterFn = j.showPods

	return &j
}

func (j *Job) extraActions(aa ui.KeyActions) {
	j.LogResource.extraActions(aa)
}

func (j *Job) showPods(app *App, _, res, sel string) {
	ns, n := namespaced(sel)
	job, err := k8s.NewJob(app.Conn()).Get(ns, n)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	jo, ok := job.(*batchv1.Job)
	if !ok {
		log.Fatal().Msg("Expecting a valid job")
	}
	l, err := metav1.LabelSelectorAsSelector(jo.Spec.Selector)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPods(app, ns, l.String(), "")
}
