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
func NewPVCList(ns string) List {
	return NewPVCListWithArgs(ns, NewPVC())
}

// NewPVCListWithArgs returns a new resource list.
func NewPVCListWithArgs(ns string, res Resource) List {
	return newList(ns, "pvc", res, AllVerbsAccess|DescribeAccess)
}

// NewPVC instantiates a new Endpoint.
func NewPVC() *PVC {
	return NewPVCWithArgs(k8s.NewPVC())
}

// NewPVCWithArgs instantiates a new Endpoint.
func NewPVCWithArgs(r k8s.Res) *PVC {
	ep := &PVC{
		Base: &Base{
			caller: r,
		},
	}
	ep.creator = ep
	return ep
}

// NewInstance builds a new Endpoint instance from a k8s resource.
func (*PVC) NewInstance(i interface{}) Columnar {
	cm := NewPVC()
	switch i.(type) {
	case *v1.PersistentVolumeClaim:
		cm.instance = i.(*v1.PersistentVolumeClaim)
	case v1.PersistentVolumeClaim:
		ii := i.(v1.PersistentVolumeClaim)
		cm.instance = &ii
	default:
		log.Fatal().Msgf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal resource to yaml.
func (r *PVC) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
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

// ExtFields returns extended fields in relation to headers.
func (*PVC) ExtFields() Properties {
	return Properties{}
}
