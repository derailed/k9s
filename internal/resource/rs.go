package resource

import (
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
)

// ReplicaSet tracks a kubernetes resource.
type ReplicaSet struct {
	*Base
	instance *v1.ReplicaSet
}

// NewReplicaSetList returns a new resource list.
func NewReplicaSetList(ns string) List {
	return NewReplicaSetListWithArgs(ns, NewReplicaSet())
}

// NewReplicaSetListWithArgs returns a new resource list.
func NewReplicaSetListWithArgs(ns string, res Resource) List {
	return newList(ns, "rs", res, AllVerbsAccess|DescribeAccess)
}

// NewReplicaSet instantiates a new Endpoint.
func NewReplicaSet() *ReplicaSet {
	return NewReplicaSetWithArgs(k8s.NewReplicaSet())
}

// NewReplicaSetWithArgs instantiates a new Endpoint.
func NewReplicaSetWithArgs(r k8s.Res) *ReplicaSet {
	ep := &ReplicaSet{
		Base: &Base{
			caller: r,
		},
	}
	ep.creator = ep
	return ep
}

// NewInstance builds a new Endpoint instance from a k8s resource.
func (*ReplicaSet) NewInstance(i interface{}) Columnar {
	cm := NewReplicaSet()
	switch i.(type) {
	case *v1.ReplicaSet:
		cm.instance = i.(*v1.ReplicaSet)
	case v1.ReplicaSet:
		ii := i.(v1.ReplicaSet)
		cm.instance = &ii
	default:
		log.Fatalf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal a deployment given a namespaced name.
func (r *ReplicaSet) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
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

// ExtFields returns extended fields in relation to headers.
func (*ReplicaSet) ExtFields() Properties {
	return Properties{}
}
