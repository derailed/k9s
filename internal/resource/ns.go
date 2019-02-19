package resource

import (
	"github.com/derailed/k9s/internal/k8s"
	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
)

// Namespace tracks a kubernetes resource.
type Namespace struct {
	*Base
	instance *v1.Namespace
}

// NewNamespaceList returns a new resource list.
func NewNamespaceList(ns string) List {
	return NewNamespaceListWithArgs(ns, NewNamespace())
}

// NewNamespaceListWithArgs returns a new resource list.
func NewNamespaceListWithArgs(ns string, res Resource) List {
	return newList(NotNamespaced, "ns", res, CRUDAccess)
}

// NewNamespace instantiates a new Endpoint.
func NewNamespace() *Namespace {
	return NewNamespaceWithArgs(k8s.NewNamespace())
}

// NewNamespaceWithArgs instantiates a new Endpoint.
func NewNamespaceWithArgs(r k8s.Res) *Namespace {
	ep := &Namespace{
		Base: &Base{
			caller: r,
		},
	}
	ep.creator = ep
	return ep
}

// NewInstance builds a new Endpoint instance from a k8s resource.
func (*Namespace) NewInstance(i interface{}) Columnar {
	cm := NewNamespace()
	switch i.(type) {
	case *v1.Namespace:
		cm.instance = i.(*v1.Namespace)
	case v1.Namespace:
		ii := i.(v1.Namespace)
		cm.instance = &ii
	default:
		log.Fatalf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal a resource to yaml.
func (r *Namespace) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		log.Error(err)
		return "", err
	}

	nss := i.(*v1.Namespace)
	nss.TypeMeta.APIVersion = "v1"
	nss.TypeMeta.Kind = "Namespace"
	return r.marshalObject(nss)
}

// Header returns resource header.
func (*Namespace) Header(ns string) Row {
	return Row{"NAME", "STATUS", "AGE"}
}

// Fields returns displayable fields.
func (r *Namespace) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance
	return append(ff,
		i.Name,
		string(i.Status.Phase),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ExtFields returns extended fields in relation to headers.
func (*Namespace) ExtFields() Properties {
	return Properties{}
}
