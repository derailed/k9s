package resource

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	v1 "k8s.io/api/apps/v1"
)

// ReplicaSet tracks a kubernetes resource.
type ReplicaSet struct {
	*Base
	instance *v1.ReplicaSet
}

// NewReplicaSetList returns a new resource list.
func NewReplicaSetList(c Connection, ns string) List {
	return NewList(
		ns,
		"rs",
		NewReplicaSet(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewReplicaSet instantiates a new ReplicaSet.
func NewReplicaSet(c Connection) *ReplicaSet {
	r := &ReplicaSet{&Base{Connection: c, Resource: k8s.NewReplicaSet(c)}, nil}
	r.Factory = r

	return r
}

// New builds a new ReplicaSet instance from a k8s resource.
func (r *ReplicaSet) New(i interface{}) (Columnar, error) {
	c := NewReplicaSet(r.Connection)
	switch instance := i.(type) {
	case *v1.ReplicaSet:
		c.instance = instance
	case v1.ReplicaSet:
		c.instance = &instance
	default:
		return nil, fmt.Errorf("Expecting ReplicaSet but got %T", instance)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c, nil
}

// Marshal a deployment given a namespaced name.
func (r *ReplicaSet) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	rs, ok := i.(*v1.ReplicaSet)
	if !ok {
		return "", errors.New("Expecting a rs resource")
	}
	rs.TypeMeta.APIVersion = "apps/v1"
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
	i := r.instance

	ff := make(Row, 0, len(r.Header(ns)))
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.Name,
		strconv.Itoa(int(*i.Spec.Replicas)),
		strconv.Itoa(int(i.Status.Replicas)),
		strconv.Itoa(int(i.Status.ReadyReplicas)),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}
