package resource

import (
	"path"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

// PV tracks a kubernetes resource.
type PV struct {
	*Base
	instance *v1.PersistentVolume
}

// NewPVList returns a new resource list.
func NewPVList(ns string) List {
	return NewPVListWithArgs(ns, NewPV())
}

// NewPVListWithArgs returns a new resource list.
func NewPVListWithArgs(ns string, res Resource) List {
	return newList(NotNamespaced, "pv", res, CRUDAccess|DescribeAccess)
}

// NewPV instantiates a new Endpoint.
func NewPV() *PV {
	return NewPVWithArgs(k8s.NewPV())
}

// NewPVWithArgs instantiates a new Endpoint.
func NewPVWithArgs(r k8s.Res) *PV {
	ep := &PV{
		Base: &Base{
			caller: r,
		},
	}
	ep.creator = ep
	return ep
}

// NewInstance builds a new Endpoint instance from a k8s resource.
func (*PV) NewInstance(i interface{}) Columnar {
	cm := NewPV()
	switch i.(type) {
	case *v1.PersistentVolume:
		cm.instance = i.(*v1.PersistentVolume)
	case v1.PersistentVolume:
		ii := i.(v1.PersistentVolume)
		cm.instance = &ii
	default:
		log.Fatal().Msgf("Unknown %#v", i)
	}
	cm.path = cm.namespacedName(cm.instance.ObjectMeta)
	return cm
}

// Marshal resource to yaml.
func (r *PV) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		return "", err
	}

	pv := i.(*v1.PersistentVolume)
	pv.TypeMeta.APIVersion = "v1"
	pv.TypeMeta.Kind = "PeristentVolume"
	return r.marshalObject(pv)
}

// Header return resource header.
func (*PV) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}
	return append(hh, "NAME", "CAPACITY", "ACCESS MODES", "RECLAIM POLICY", "STATUS", "CLAIM", "STORAGECLASS", "REASON", "AGE")
}

// Fields retrieves displayable fields.
func (r *PV) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance
	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	phase := i.Status.Phase
	if i.ObjectMeta.DeletionTimestamp != nil {
		phase = "Terminating"
	}

	var claim string
	if i.Spec.ClaimRef != nil {
		claim = path.Join(i.Spec.ClaimRef.Namespace, i.Spec.ClaimRef.Name)
	}

	class, found := i.Annotations[v1.BetaStorageClassAnnotation]
	if !found {
		class = i.Spec.StorageClassName
	}

	size := i.Spec.Capacity[v1.ResourceStorage]

	return append(ff,
		i.Name,
		size.String(),
		r.accessMode(i.Spec.AccessModes),
		string(i.Spec.PersistentVolumeReclaimPolicy),
		string(phase),
		claim,
		class,
		i.Status.Reason,
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ExtFields returns extended fields in relation to headers.
func (*PV) ExtFields() Properties {
	return Properties{}
}

// Helpers...

func (r *PV) accessMode(aa []v1.PersistentVolumeAccessMode) string {
	dd := r.accessDedup(aa)
	s := make([]string, 0, len(dd))
	for i := 0; i < len(aa); i++ {
		switch {
		case r.accessContains(dd, v1.ReadWriteOnce):
			s = append(s, "RWO")
		case r.accessContains(dd, v1.ReadOnlyMany):
			s = append(s, "ROX")
		case r.accessContains(dd, v1.ReadWriteMany):
			s = append(s, "RWX")
		}
	}
	return strings.Join(s, ",")
}

func (r *PV) accessContains(cc []v1.PersistentVolumeAccessMode, a v1.PersistentVolumeAccessMode) bool {
	for _, c := range cc {
		if c == a {
			return true
		}
	}
	return false
}

func (r *PV) accessDedup(cc []v1.PersistentVolumeAccessMode) []v1.PersistentVolumeAccessMode {
	set := []v1.PersistentVolumeAccessMode{}
	for _, c := range cc {
		if !r.accessContains(set, c) {
			set = append(set, c)
		}
	}
	return set
}
