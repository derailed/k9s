package resource

import (
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/apps/v1"
)

// StatefulSet tracks a kubernetes resource.
type StatefulSet struct {
	*Base
	instance *v1.StatefulSet
}

// NewStatefulSetList returns a new resource list.
func NewStatefulSetList(c Connection, ns string) List {
	return NewList(
		ns,
		"sts",
		NewStatefulSet(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewStatefulSet instantiates a new StatefulSet.
func NewStatefulSet(c Connection) *StatefulSet {
	s := &StatefulSet{&Base{Connection: c, Resource: k8s.NewStatefulSet(c)}, nil}
	s.Factory = s

	return s
}

// New builds a new StatefulSet instance from a k8s resource.
func (r *StatefulSet) New(i interface{}) Columnar {
	c := NewStatefulSet(r.Connection)
	switch instance := i.(type) {
	case *v1.StatefulSet:
		c.instance = instance
	case v1.StatefulSet:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown StatefulSet type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *StatefulSet) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	sts := i.(*v1.StatefulSet)
	sts.TypeMeta.APIVersion = "v1"
	sts.TypeMeta.Kind = "StatefulSet"

	return r.marshalObject(sts)
}

// Header return resource header.
func (*StatefulSet) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}

	return append(hh, "NAME", "DESIRED", "CURRENT", "AGE")
}

// Fields retrieves displayable fields.
func (r *StatefulSet) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.Name,
		strconv.Itoa(int(*i.Spec.Replicas)),
		strconv.Itoa(int(i.Status.ReadyReplicas)),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}
