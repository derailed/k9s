package resource

import (
	"context"
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
)

// DaemonSet tracks a kubernetes resource.
type DaemonSet struct {
	*Base
	instance *appsv1.DaemonSet
}

// NewDaemonSetList returns a new resource list.
func NewDaemonSetList(c Connection, ns string) List {
	return NewList(
		ns,
		"ds",
		NewDaemonSet(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewDaemonSet instantiates a new DaemonSet.
func NewDaemonSet(c Connection) *DaemonSet {
	ds := &DaemonSet{&Base{Connection: c, Resource: k8s.NewDaemonSet(c)}, nil}
	ds.Factory = ds

	return ds
}

// New builds a new DaemonSet instance from a k8s resource.
func (r *DaemonSet) New(i interface{}) Columnar {
	c := NewDaemonSet(r.Connection)
	switch instance := i.(type) {
	case *appsv1.DaemonSet:
		c.instance = instance
	case appsv1.DaemonSet:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown DaemonSet type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *DaemonSet) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	ds := i.(*appsv1.DaemonSet)
	ds.TypeMeta.APIVersion = "apps/v1"
	ds.TypeMeta.Kind = "DaemonSet"

	return r.marshalObject(ds)
}

// Logs tail logs for all pods represented by this DaemonSet.
func (r *DaemonSet) Logs(ctx context.Context, c chan<- string, opts LogOptions) error {
	instance, err := r.Resource.Get(opts.Namespace, opts.Name)
	if err != nil {
		return err
	}

	ds := instance.(*appsv1.DaemonSet)
	if ds.Spec.Selector == nil || len(ds.Spec.Selector.MatchLabels) == 0 {
		return fmt.Errorf("No valid selector found on daemonset %s", opts.FQN())
	}

	return r.podLogs(ctx, c, ds.Spec.Selector.MatchLabels, opts)
}

// Header return resource header.
func (*DaemonSet) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}
	hh = append(hh, "NAME", "DESIRED", "CURRENT", "READY", "UP-TO-DATE")
	hh = append(hh, "AVAILABLE", "NODE_SELECTOR", "AGE")

	return hh
}

// Fields retrieves displayable fields.
func (r *DaemonSet) Fields(ns string) Row {
	ff := make([]string, 0, len(r.Header(ns)))

	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.Name,
		strconv.Itoa(int(i.Status.DesiredNumberScheduled)),
		strconv.Itoa(int(i.Status.CurrentNumberScheduled)),
		strconv.Itoa(int(i.Status.NumberReady)),
		strconv.Itoa(int(i.Status.UpdatedNumberScheduled)),
		strconv.Itoa(int(i.Status.NumberAvailable)),
		mapToStr(i.Spec.Template.Spec.NodeSelector),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}
