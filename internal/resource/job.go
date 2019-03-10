package resource

import (
	"bufio"
	"context"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

// Job tracks a kubernetes resource.
type Job struct {
	*Base
	instance *v1.Job
}

// NewJobList returns a new resource list.
func NewJobList(ns string) List {
	return NewJobListWithArgs(ns, NewJob())
}

// NewJobListWithArgs returns a new resource list.
func NewJobListWithArgs(ns string, res Resource) List {
	return newList(ns, "job", res, AllVerbsAccess|DescribeAccess)
}

// NewJob instantiates a new Job.
func NewJob() *Job {
	return NewJobWithArgs(k8s.NewJob())
}

// NewJobWithArgs instantiates a new Job.
func NewJobWithArgs(r k8s.Res) *Job {
	cm := &Job{
		Base: &Base{
			caller: r,
		},
	}
	cm.creator = cm
	return cm
}

// NewInstance builds a new Job instance from a k8s resource.
func (*Job) NewInstance(i interface{}) Columnar {
	job := NewJob()
	switch i.(type) {
	case *v1.Job:
		job.instance = i.(*v1.Job)
	case v1.Job:
		ii := i.(v1.Job)
		job.instance = &ii
	default:
		log.Fatal().Msgf("Unknown %#v", i)
	}
	job.path = job.namespacedName(job.instance.ObjectMeta)
	return job
}

// Marshal resource to yaml.
func (r *Job) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		return "", err
	}

	jo := i.(*v1.Job)
	jo.TypeMeta.APIVersion = "extensions/v1beta1"
	jo.TypeMeta.Kind = "Job"
	return r.marshalObject(jo)
}

// Containers fetch all the containers on this job, may include init containers.
func (r *Job) Containers(path string, includeInit bool) ([]string, error) {
	ns, n := namespaced(path)
	return r.caller.(k8s.Loggable).Containers(ns, n, includeInit)
}

// Logs retrieves logs for a given container.
func (r *Job) Logs(c chan<- string, ns, n, co string, lines int64, prev bool) (context.CancelFunc, error) {
	req := r.caller.(k8s.Loggable).Logs(ns, n, co, lines, prev)
	ctx, cancel := context.WithCancel(context.TODO())
	req.Context(ctx)

	blocked := true
	go func() {
		select {
		case <-time.After(defaultTimeout):
			if blocked {
				close(c)
				cancel()
			}
		}
	}()

	// This call will block if nothing is in the stream!!
	stream, err := req.Stream()
	blocked = false
	if err != nil {
		return cancel, fmt.Errorf("Log tail request failed for job `%s/%s:%s", ns, n, co)
	}

	go func() {
		defer func() {
			stream.Close()
			cancel()
			close(c)
		}()

		scanner := bufio.NewScanner(stream)
		for scanner.Scan() {
			c <- scanner.Text()
		}
	}()
	return cancel, nil
}

// Header return resource header.
func (*Job) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}
	return append(hh, "NAME", "COMPLETIONS", "DURATION", "AGE")
}

// Fields retrieves displayable fields.
func (r *Job) Fields(ns string) Row {
	ff := make([]string, 0, len(r.Header(ns)))

	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}
	return append(ff,
		i.Name,
		r.toCompletion(i.Spec, i.Status),
		r.toDuration(i.Status),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ExtFields returns extended fields in relation to headers.
func (*Job) ExtFields() Properties {
	return Properties{}
}

// Helpers...

func (*Job) toCompletion(spec v1.JobSpec, status v1.JobStatus) (s string) {
	if spec.Completions != nil {
		return fmt.Sprintf("%d/%d", status.Succeeded, *spec.Completions)
	}
	var parallelism int32
	if spec.Parallelism != nil {
		parallelism = *spec.Parallelism
	}
	if parallelism > 1 {
		return fmt.Sprintf("%d/1 of %d", status.Succeeded, parallelism)
	}
	return fmt.Sprintf("%d/1", status.Succeeded)
}

func (*Job) toDuration(status v1.JobStatus) string {
	switch {
	case status.StartTime == nil:
	case status.CompletionTime == nil:
		return duration.HumanDuration(time.Since(status.StartTime.Time))
	}
	return duration.HumanDuration(status.CompletionTime.Sub(status.StartTime.Time))
}
