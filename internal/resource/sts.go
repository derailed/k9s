package resource

import (
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/apps/v1"
)

// StatefulSet tracks a kubernetes resource.
type StatefulSet struct {
	*Base
	instance *v1.StatefulSet
}

// NewStatefulSetList returns a new resource list.
func NewStatefulSetList(ns string) List {
	return NewStatefulSetListWithArgs(ns, NewStatefulSet())
}

// NewStatefulSetListWithArgs returns a new resource list.
func NewStatefulSetListWithArgs(ns string, res Resource) List {
	return newList(ns, "sts", res, AllVerbsAccess|DescribeAccess)
}

// NewStatefulSet instantiates a new Endpoint.
func NewStatefulSet() *StatefulSet {
	return NewStatefulSetWithArgs(k8s.NewStatefulSet())
}

// NewStatefulSetWithArgs instantiates a new Endpoint.
func NewStatefulSetWithArgs(r k8s.Res) *StatefulSet {
	ep := &StatefulSet{
		Base: &Base{
			caller: r,
		},
	}
	ep.creator = ep
	return ep
}

// NewInstance builds a new Endpoint instance from a k8s resource.
func (*StatefulSet) NewInstance(i interface{}) Columnar {
	cm := NewStatefulSet()
	switch i.(type) {
	case *v1.StatefulSet:
		cm.instance = i.(*v1.StatefulSet)
	case v1.StatefulSet:
		ii := i.(v1.StatefulSet)
		cm.instance = &ii
	default:
		log.Fatalf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal resource to yaml.
func (r *StatefulSet) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
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

// ExtFields returns extended fields in relation to headers.
func (*StatefulSet) ExtFields() Properties {
	return Properties{}
}
