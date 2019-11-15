package resource

import (
	"errors"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

// PersistentVolumeClaim tracks a kubernetes resource.
type PersistentVolumeClaim struct {
	*Base
	instance *v1.PersistentVolumeClaim
}

// NewPersistentVolumeClaimList returns a new resource list.
func NewPersistentVolumeClaimList(c Connection, ns string) List {
	return NewList(
		ns,
		"pvc",
		NewPersistentVolumeClaim(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewPersistentVolumeClaim instantiates a new PersistentVolumeClaim.
func NewPersistentVolumeClaim(c Connection) *PersistentVolumeClaim {
	p := &PersistentVolumeClaim{&Base{Connection: c, Resource: k8s.NewPersistentVolumeClaim(c)}, nil}
	p.Factory = p

	return p
}

// New builds a new PersistentVolumeClaim instance from a k8s resource.
func (r *PersistentVolumeClaim) New(i interface{}) Columnar {
	c := NewPersistentVolumeClaim(r.Connection)
	switch instance := i.(type) {
	case *v1.PersistentVolumeClaim:
		c.instance = instance
	case v1.PersistentVolumeClaim:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown PersistentVolumeClaim type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *PersistentVolumeClaim) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	pvc, ok := i.(*v1.PersistentVolumeClaim)
	if !ok {
		return "", errors.New("Expecting a pvc resource")
	}
	pvc.TypeMeta.APIVersion = "v1"
	pvc.TypeMeta.Kind = "PersistentVolumeClaim"

	return r.marshalObject(pvc)
}

// Header return resource header.
func (*PersistentVolumeClaim) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}

	return append(hh, "NAME", "STATUS", "VOLUME", "CAPACITY", "ACCESS MODES", "STORAGECLASS", "AGE")
}

// Fields retrieves displayable fields.
func (r *PersistentVolumeClaim) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	phase := i.Status.Phase
	if i.ObjectMeta.DeletionTimestamp != nil {
		phase = Terminating
	}

	var pv PersistentVolume
	storage := i.Spec.Resources.Requests[v1.ResourceStorage]
	var capacity, accessModes string
	if i.Spec.VolumeName != "" {
		accessModes = pv.accessMode(i.Status.AccessModes)
		storage = i.Status.Capacity[v1.ResourceStorage]
		capacity = storage.String()
	}

	class, found := i.Annotations[v1.BetaStorageClassAnnotation]
	if !found {
		if i.Spec.StorageClassName != nil {
			class = *i.Spec.StorageClassName
		}
	}

	return append(ff,
		i.Name,
		string(phase),
		i.Spec.VolumeName,
		capacity,
		accessModes,
		class,
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}
