package resource

import (
	"errors"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

// Namespace tracks a kubernetes resource.
type Namespace struct {
	*Base
	instance *v1.Namespace
}

// NewNamespaceList returns a new resource list.
func NewNamespaceList(c Connection, ns string) List {
	return NewList(
		NotNamespaced,
		"ns",
		NewNamespace(c),
		CRUDAccess|DescribeAccess,
	)
}

// NewNamespace instantiates a new Namespace.
func NewNamespace(c Connection) *Namespace {
	n := &Namespace{&Base{Connection: c, Resource: k8s.NewNamespace(c)}, nil}
	n.Factory = n

	return n
}

// New builds a new Namespace instance from a k8s resource.
func (r *Namespace) New(i interface{}) Columnar {
	c := NewNamespace(r.Connection)
	switch instance := i.(type) {
	case *v1.Namespace:
		c.instance = instance
	case v1.Namespace:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown Namespace type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal a resource to yaml.
func (r *Namespace) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		log.Error().Err(err)
		return "", err
	}

	nss, ok := i.(*v1.Namespace)
	if !ok {
		return "", errors.New("Expecting a ns resource")
	}
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
