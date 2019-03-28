package resource

import (
	"bufio"
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/duration"
)

// Job tracks a kubernetes resource.
type Job struct {
	*Base
	instance *batchv1.Job
}

// NewJobList returns a new resource list.
func NewJobList(c Connection, ns string) List {
	return NewList(
		ns,
		"job",
		NewJob(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewJob instantiates a new Job.
func NewJob(c Connection) *Job {
	j := &Job{&Base{Connection: c, Resource: k8s.NewJob(c)}, nil}
	j.Factory = j

	return j
}

// New builds a new Job instance from a k8s resource.
func (r *Job) New(i interface{}) Columnar {
	c := NewJob(r.Connection)
	switch instance := i.(type) {
	case *batchv1.Job:
		c.instance = instance
	case batchv1.Job:
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
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	jo := i.(*batchv1.Job)
	jo.TypeMeta.APIVersion = "extensions/v1beta1"
	jo.TypeMeta.Kind = "Job"

	return r.marshalObject(jo)
}

// Containers fetch all the containers on this job, may include init containers.
func (r *Job) Containers(path string, includeInit bool) ([]string, error) {
	ns, n := namespaced(path)

	return r.Resource.(k8s.Loggable).Containers(ns, n, includeInit)
}

// Logs retrieves logs for a given container.
func (r *Job) Logs(c chan<- string, ns, n, co string, lines int64, prev bool) (context.CancelFunc, error) {
	req := r.Resource.(k8s.Loggable).Logs(ns, n, co, lines, prev)
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

	return append(hh, "NAME", "COMPLETIONS", "DURATION", "CONTAINERS", "IMAGES", "AGE")
}

// Fields retrieves displayable fields.
func (r *Job) Fields(ns string) Row {
	ff := make([]string, 0, len(r.Header(ns)))

	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	cc, ii := r.toContainers(i.Spec.Template.Spec)

	return append(ff,
		i.Name,
		r.toCompletion(i.Spec, i.Status),
		r.toDuration(i.Status),
		cc,
		ii,
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ----------------------------------------------------------------------------
// Helpers...

func (*Job) toContainers(p v1.PodSpec) (string, string) {
	cc := make([]string, 0, len(p.InitContainers)+len(p.Containers))
	ii := make([]string, 0, len(cc))

	for _, c := range p.InitContainers {
		cc = append(cc, c.Name)
		ii = append(ii, c.Image)
	}
	for _, c := range p.Containers {
		cc = append(cc, c.Name)
		ii = append(ii, c.Image)
	}

	const maxShow = 2
	// Limit to 2 of each...
	if len(cc) > maxShow {
		cc = append(cc[:2], fmt.Sprintf("(+%d)...", len(cc)-maxShow))
	}
	if len(ii) > maxShow {
		ii = append(ii[:2], fmt.Sprintf("(+%d)...", len(ii)-maxShow))
	}

	return strings.Join(cc, ","), strings.Join(ii, ",")
}

func (*Job) toCompletion(spec batchv1.JobSpec, status batchv1.JobStatus) (s string) {
	if spec.Completions != nil {
		return fmt.Sprintf("%d/%d", status.Succeeded, *spec.Completions)
	}

	if spec.Parallelism == nil {
		return fmt.Sprintf("%d/1", status.Succeeded)
	}

	p := *spec.Parallelism
	if p > 1 {
		return fmt.Sprintf("%d/1 of %d", status.Succeeded, p)
	}

	return fmt.Sprintf("%d/1", status.Succeeded)
}

func (*Job) toDuration(status batchv1.JobStatus) string {
	if status.StartTime == nil {
		return MissingValue
	}

	var d time.Duration
	switch {
	case status.CompletionTime == nil:
		d = time.Since(status.StartTime.Time)
	default:
		d = status.CompletionTime.Sub(status.StartTime.Time)
	}

	return duration.HumanDuration(d)
}
