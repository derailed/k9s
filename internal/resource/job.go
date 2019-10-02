package resource

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
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
	mx       sync.RWMutex
}

// NewJobList returns a new resource list.
func NewJobList(c Connection, ns string, gvr k8s.GVR) List {
	return NewList(
		ns,
		"job",
		NewJob(c, gvr),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewJob instantiates a new Job.
func NewJob(c Connection, gvr k8s.GVR) *Job {
	j := &Job{
		Base: &Base{Connection: c, Resource: k8s.NewJob(c, gvr)},
	}
	j.Factory = j

	return j
}

// New builds a new Job instance from a k8s resource.
func (r *Job) New(i interface{}) Columnar {
	c := NewJob(r.Connection, r.GVR())
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
	ns, n := Namespaced(path)
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
	ns, n := Namespaced(path)

	return r.Resource.(k8s.Loggable).Containers(ns, n, includeInit)
}

// Logs retrieves logs for a given container.
func (r *Job) Logs(ctx context.Context, c chan<- string, opts LogOptions) error {
	instance, err := r.Resource.Get(opts.Namespace, opts.Name)
	if err != nil {
		return err
	}
	jo := instance.(*batchv1.Job)
	if jo.Spec.Selector == nil || len(jo.Spec.Selector.MatchLabels) == 0 {
		return fmt.Errorf("No valid selector found on job %s", opts.FQN())
	}

	return r.podLogs(ctx, c, jo.Spec.Selector.MatchLabels, opts)
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

const maxShow = 2

func (*Job) toContainers(p v1.PodSpec) (string, string) {
	cc, ii := parseContainers(p.InitContainers)
	cn, ci := parseContainers(p.Containers)

	cc, ii = append(cc, cn...), append(ii, ci...)

	// Limit to 2 of each...
	if len(cc) > maxShow {
		cc = append(cc[:2], "(+"+strconv.Itoa(len(cc)-maxShow)+")...")
	}
	if len(ii) > maxShow {
		ii = append(ii[:2], "(+"+strconv.Itoa(len(ii)-maxShow)+")...")
	}

	return strings.Join(cc, ","), strings.Join(ii, ",")
}

func parseContainers(cos []v1.Container) (nn, ii []string) {
	for _, co := range cos {
		nn = append(nn, co.Name)
		ii = append(ii, co.Image)
	}

	return nn, ii
}

func (*Job) toCompletion(spec batchv1.JobSpec, status batchv1.JobStatus) (s string) {
	if spec.Completions != nil {
		return strconv.Itoa(int(status.Succeeded)) + "/" + strconv.Itoa(int(*spec.Completions))
	}

	if spec.Parallelism == nil {
		return strconv.Itoa(int(status.Succeeded)) + "/1"
	}

	p := *spec.Parallelism
	if p > 1 {
		return strconv.Itoa(int(status.Succeeded)) + "/1 of " + strconv.Itoa(int(p))
	}

	return strconv.Itoa(int(status.Succeeded)) + "/1"
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
