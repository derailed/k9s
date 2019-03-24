package resource

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

// PVC tracks a kubernetes resource.
type PVC struct {
	*Base
	instance *v1.PersistentVolumeClaim
}

// NewPVCList returns a new resource list.
func NewPVCList(c k8s.Connection, ns string) List {
	return newList(
		ns,
		"pvc",
		NewPVC(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewPVC instantiates a new PVC.
func NewPVC(c k8s.Connection) *PVC {
	p := &PVC{&Base{connection: c, resource: k8s.NewPVC(c)}, nil}
	p.Factory = p

	return p
}

// New builds a new PVC instance from a k8s resource.
func (r *PVC) New(i interface{}) Columnar {
	c := NewPVC(r.connection)
	switch instance := i.(type) {
	case *v1.PersistentVolumeClaim:
		c.instance = instance
	case v1.PersistentVolumeClaim:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown PVC type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *PVC) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	pvc := i.(*v1.PersistentVolumeClaim)
	pvc.TypeMeta.APIVersion = "v1"
	pvc.TypeMeta.Kind = "PersistentVolumeClaim"

	return r.marshalObject(pvc)
}

// Header return resource header.
func (*PVC) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}

	return append(hh, "NAME", "STATUS", "VOLUME", "CAPACITY", "ACCESS MODES", "STORAGECLASS", "AGE")
}

// Fields retrieves displayable fields.
func (r *PVC) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	phase := i.Status.Phase
	if i.ObjectMeta.DeletionTimestamp != nil {
		phase = "Terminating"
	}

	pv := PV{}
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
