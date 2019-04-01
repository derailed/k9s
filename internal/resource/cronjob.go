package resource

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
)

type (
	// CronJob tracks a kubernetes resource.
	CronJob struct {
		*Base
		instance *batchv1beta1.CronJob
	}

	// Runner can run jobs.
	Runner interface {
		Run(path string) error
	}

	// Runnable can run jobs.
	Runnable interface {
		Run(ns, n string) error
	}
)

// NewCronJobList returns a new resource list.
func NewCronJobList(c Connection, ns string) List {
	return NewList(
		ns,
		"cronjob",
		NewCronJob(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewCronJob instantiates a new CronJob.
func NewCronJob(c Connection) *CronJob {
	cj := &CronJob{&Base{Connection: c, Resource: k8s.NewCronJob(c)}, nil}
	cj.Factory = cj

	return cj
}

// New builds a new CronJob instance from a k8s resource.
func (r *CronJob) New(i interface{}) Columnar {
	c := NewCronJob(r.Connection)
	switch instance := i.(type) {
	case *batchv1beta1.CronJob:
		c.instance = instance
	case batchv1beta1.CronJob:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown CronJob type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *CronJob) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	cj := i.(*batchv1beta1.CronJob)
	cj.TypeMeta.APIVersion = "extensions/batchv1beta1"
	cj.TypeMeta.Kind = "CronJob"

	return r.marshalObject(cj)
}

// Run a given cronjob.
func (r *CronJob) Run(pa string) error {
	ns, n := namespaced(pa)
	if c, ok := r.Resource.(Runnable); ok {
		return c.Run(ns, n)
	}

	return fmt.Errorf("unable to run cronjob %s", pa)
}

// Header return resource header.
func (*CronJob) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}

	return append(hh, "NAME", "SCHEDULE", "SUSPEND", "ACTIVE", "LAST_SCHEDULE", "AGE")
}

// Fields retrieves displayable fields.
func (r *CronJob) Fields(ns string) Row {
	ff := make([]string, 0, len(r.Header(ns)))

	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	lastScheduled := "<none>"
	if i.Status.LastScheduleTime != nil {
		lastScheduled = toAgeHuman(toAge(*i.Status.LastScheduleTime))
	}

	return append(ff,
		i.Name,
		i.Spec.Schedule,
		boolPtrToStr(i.Spec.Suspend),
		strconv.Itoa(len(i.Status.Active)),
		lastScheduled,
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}
