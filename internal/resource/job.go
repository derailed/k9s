package resource

import (
	"fmt"
	"time"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
	"k8s.io/api/batch/v1"
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
	return newList(ns, "job", res, AllVerbsAccess)
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
		log.Fatalf("Unknown %#v", i)
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

	dp := i.(*v1.Job)
	dp.TypeMeta.APIVersion = "extensions/v1beta1"
	dp.TypeMeta.Kind = "Job"
	raw, err := yaml.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(raw), nil
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
