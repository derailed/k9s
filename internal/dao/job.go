package dao

import (
	"context"
	"errors"
	"fmt"

	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// Job represents a K8s job resource.
type Job struct {
	Generic
}

var _ Accessor = (*Job)(nil)
var _ Loggable = (*Job)(nil)

// TailLogs tail logs for all pods represented by this Job.
func (j *Job) TailLogs(ctx context.Context, c chan<- string, opts LogOptions) error {
	o, err := j.Get(j.gvr.String(), opts.Path, true, labels.Everything())
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
