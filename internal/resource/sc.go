package resource

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/storage/v1"
)

// StorageClass tracks a kubernetes resource.
type StorageClass struct {
	*Base
	instance *v1.StorageClass
}

// NewStorageClassList returns a new resource list.
func NewStorageClassList(c Connection, ns string, gvr k8s.GVR) List {
	return NewList(
		NotNamespaced,
		"sc",
		NewStorageClass(c, gvr),
		CRUDAccess|DescribeAccess,
	)
}

// NewStorageClass instantiates a new StorageClass.
func NewStorageClass(c Connection, gvr k8s.GVR) *StorageClass {
	p := &StorageClass{&Base{Connection: c, Resource: k8s.NewStorageClass(c, gvr)}, nil}
	p.Factory = p

	return p
}

// New builds a new StorageClass instance from a k8s resource.
func (r *StorageClass) New(i interface{}) Columnar {
	c := NewStorageClass(r.Connection, k8s.GVR{})
	switch instance := i.(type) {
	case *v1.StorageClass:
		c.instance = instance
	case v1.StorageClass:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown StorageClass type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *StorageClass) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	sc := i.(*v1.StorageClass)
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
