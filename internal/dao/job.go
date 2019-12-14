package dao

import (
	"context"
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

type Job struct {
	Generic
}

var _ Accessor = &Job{}
var _ Loggable = &Job{}

// Logs tail logs for all pods represented by this Job.
func (j *Job) TailLogs(ctx context.Context, c chan<- string, opts LogOptions) error {
	log.Debug().Msgf("Tailing Job %#v", opts)
	o, err := j.Get(string(j.gvr), opts.Path, labels.Everything())
	if err != nil {
		return err
	}

	var job batchv1.Job
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &job)
	if err != nil {
		return errors.New("expecting a job resource")
	}

	if job.Spec.Selector == nil || len(job.Spec.Selector.MatchLabels) == 0 {
		return fmt.Errorf("No valid selector found on Job %s", opts.Path)
	}

	return podLogs(ctx, c, job.Spec.Selector.MatchLabels, opts)
}
