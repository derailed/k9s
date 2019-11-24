package view

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
)

// Job represents a job viewer.
type Job struct {
	ResourceViewer
}

// NewJob returns a new viewer.
func NewJob(title, gvr string, list resource.List) ResourceViewer {
	j := Job{
		ResourceViewer: NewLogsExtender(
			NewResource(title, gvr, list),
			func() string { return "" },
		),
	}
	j.GetTable().SetEnterFn(j.showPods)

	return &j
}

func (j *Job) showPods(app *App, _, res, path string) {
	ns, n := namespaced(path)
	job, err := k8s.NewJob(app.Conn()).Get(ns, n)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	jo, ok := job.(*batchv1.Job)
	if !ok {
		log.Fatal().Msg("Expecting a valid job")
	}
	showPodsFromSelector(app, ns, jo.Spec.Selector)
}
