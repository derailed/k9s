package resource

import (
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal/k8s"
	v1 "k8s.io/api/storage/v1"
)

// StorageClass tracks a kubernetes resource.
type StorageClass struct {
	*Base
	instance *v1.StorageClass
}

// NewStorageClassList returns a new resource list.
func NewStorageClassList(c Connection, ns string) List {
	return NewList(
		NotNamespaced,
		"sc",
		NewStorageClass(c),
		CRUDAccess|DescribeAccess,
	)
}

// NewStorageClass instantiates a new StorageClass.
func NewStorageClass(c Connection) *StorageClass {
	p := &StorageClass{&Base{Connection: c, Resource: k8s.NewStorageClass(c)}, nil}
	p.Factory = p

	return p
}

// New builds a new StorageClass instance from a k8s resource.
func (r *StorageClass) New(i interface{}) (Columnar, error) {
	c := NewStorageClass(r.Connection)
	switch instance := i.(type) {
	case *v1.StorageClass:
		c.instance = instance
	case v1.StorageClass:
		c.instance = &instance
	default:
		return nil, fmt.Errorf("Expecting StorageClass but got %T", instance)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c, nil
}

// Marshal resource to yaml.
func (r *StorageClass) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	sc, ok := i.(*v1.StorageClass)
	if !ok {
		return "", errors.New("Expecting a sc resource")
	}
	sc.TypeMeta.APIVersion = "storage.k8s.io/v1"
	sc.TypeMeta.Kind = "StorageClass"

	return r.marshalObject(sc)
}

// Header return resource header.
func (*StorageClass) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}

	return append(hh, "NAME", "PROVISIONER", "AGE")
}

// Fields retrieves displayable fields.
func (r *StorageClass) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	return append(ff,
		i.Name,
		string(i.Provisioner),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}
