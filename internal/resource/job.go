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
func NewJobList(c k8s.Connection, ns string) List {
	return newList(
		ns,
		"job",
		NewJob(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewJob instantiates a new Job.
func NewJob(c k8s.Connection) *Job {
	j := &Job{&Base{connection: c, resource: k8s.NewJob(c)}, nil}
	j.Factory = j

	return j
}

// New builds a new Job instance from a k8s resource.
func (r *Job) New(i interface{}) Columnar {
	c := NewJob(r.connection)
	switch instance := i.(type) {
	case *v1.Job:
		c.instance = instance
	case v1.Job:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown Job type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *Job) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.resource.Get(ns, n)
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

	return r.resource.(k8s.Loggable).Containers(ns, n, includeInit)
}

// Logs retrieves logs for a given container.
func (r *Job) Logs(c chan<- string, ns, n, co string, lines int64, prev bool) (context.CancelFunc, error) {
	req := r.resource.(k8s.Loggable).Logs(ns, n, co, lines, prev)
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

// ----------------------------------------------------------------------------
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
