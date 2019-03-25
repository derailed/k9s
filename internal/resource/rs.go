package resource

import (
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/apps/v1"
)

// ReplicaSet tracks a kubernetes resource.
type ReplicaSet struct {
	*Base
	instance *v1.ReplicaSet
}

// NewReplicaSetList returns a new resource list.
func NewReplicaSetList(c k8s.Connection, ns string) List {
	return NewList(
		ns,
		"rs",
		NewReplicaSet(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewReplicaSet instantiates a new ReplicaSet.
func NewReplicaSet(c k8s.Connection) *ReplicaSet {
	r := &ReplicaSet{&Base{Connection: c, Resource: k8s.NewReplicaSet(c)}, nil}
	r.Factory = r

	return r
}

// New builds a new ReplicaSet instance from a k8s resource.
func (r *ReplicaSet) New(i interface{}) Columnar {
	c := NewReplicaSet(r.Connection)
	switch instance := i.(type) {
	case *v1.ReplicaSet:
		c.instance = instance
	case v1.ReplicaSet:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown ReplicaSet type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal a deployment given a namespaced name.
func (r *ReplicaSet) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	rs := i.(*v1.ReplicaSet)
	rs.TypeMeta.APIVersion = "extensions/v1beta"
	rs.TypeMeta.Kind = "ReplicaSet"

	return r.marshalObject(rs)
}

// Header return resource header.
func (*ReplicaSet) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}

	return append(hh, "NAME", "DESIRED", "CURRENT", "READY", "AGE")
}

// Fields retrieves displayable fields.
func (r *ReplicaSet) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	if ns == AllNamespaces {
		ff = append(ff, r.instance.Namespace)
	}

	i := r.instance

	return append(ff,
		i.Name,
		strconv.Itoa(int(*i.Spec.Replicas)),
		strconv.Itoa(int(i.Status.Replicas)),
		strconv.Itoa(int(i.Status.ReadyReplicas)),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}
