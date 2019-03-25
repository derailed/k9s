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
func NewPVList(c k8s.Connection, ns string) List {
	return NewList(
		NotNamespaced,
		"pv",
		NewPV(c),
		CRUDAccess|DescribeAccess,
	)
}

// NewPV instantiates a new PV.
func NewPV(c k8s.Connection) *PV {
	p := &PV{&Base{Connection: c, Resource: k8s.NewPV(c)}, nil}
	p.Factory = p

	return p
}

// New builds a new PV instance from a k8s resource.
func (r *PV) New(i interface{}) Columnar {
	c := NewPV(r.Connection)
	switch instance := i.(type) {
	case *v1.PersistentVolume:
		c.instance = instance
	case v1.PersistentVolume:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown PV type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *PV) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
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

// ----------------------------------------------------------------------------
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
