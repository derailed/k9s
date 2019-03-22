package resource

import (
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1beta1 "k8s.io/api/policy/v1beta1"
)

// PodDisruptionBudget that can be displayed in a table and interacted with.
type PodDisruptionBudget struct {
	*Base
	instance *v1beta1.PodDisruptionBudget
}

// NewPDBList returns a new resource list.
func NewPDBList(ns string) List {
	return NewPDBListWithArgs(ns, NewPDB())
}

// NewPDBListWithArgs returns a new resource list.
func NewPDBListWithArgs(ns string, res Resource) List {
	return newList(ns, "pdb", res, AllVerbsAccess|DescribeAccess)
}

// NewPDB returns a new PodDisruptionBudget instance.
func NewPDB() *PodDisruptionBudget {
	return NewPDBWithArgs(k8s.NewPodDisruptionBudget())
}

// NewPDBWithArgs returns a new Pod instance.
func NewPDBWithArgs(r k8s.Res) *PodDisruptionBudget {
	p := &PodDisruptionBudget{
		Base: &Base{
			caller: r,
		},
	}
	p.creator = p
	return p
}

// NewInstance builds a new PodDisruptionBudget instance from a k8s resource.
func (r *PodDisruptionBudget) NewInstance(i interface{}) Columnar {
	pdb := NewPDB()
	switch i.(type) {
	case *v1beta1.PodDisruptionBudget:
		pdb.instance = i.(*v1beta1.PodDisruptionBudget)
	case v1beta1.PodDisruptionBudget:
		ii := i.(v1beta1.PodDisruptionBudget)
		pdb.instance = &ii
	case *interface{}:
		ptr := *i.(*interface{})
		pdbi := ptr.(v1beta1.PodDisruptionBudget)
		pdb.instance = &pdbi
	default:
		log.Fatal().Msgf("Unknown %#v", i)
	}
	pdb.path = r.namespacedName(pdb.instance.ObjectMeta)
	return pdb
}

// Marshal resource to yaml.
func (r *PodDisruptionBudget) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.caller.Get(ns, n)
	if err != nil {
		return "", err
	}

	pdb := i.(*v1beta1.PodDisruptionBudget)
	pdb.TypeMeta.APIVersion = "v1beta1"
	pdb.TypeMeta.Kind = "PodDisruptionBudget"
	return r.marshalObject(pdb)
}

// Header return resource header.
func (*PodDisruptionBudget) Header(ns string) Row {
	hh := Row{}
	if ns == AllNamespaces {
		hh = append(hh, "NAMESPACE")
	}
	return append(hh,
		"NAME",
		"MIN AVAILABLE",
		"MAX_ UNAVAILABLE",
		"ALLOWED DISRUPTIONS",
		"CURRENT",
		"DESIRED",
		"EXPECTED",
		"AGE",
	)
}

// Fields retrieves displayable fields.
func (r *PodDisruptionBudget) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))
	i := r.instance

	if ns == AllNamespaces {
		ff = append(ff, i.Namespace)
	}

	min := NAValue
	if i.Spec.MinAvailable != nil {
		min = strconv.Itoa(int(i.Spec.MinAvailable.IntVal))
	}

	max := NAValue
	if i.Spec.MaxUnavailable != nil {
		max = strconv.Itoa(int(i.Spec.MaxUnavailable.IntVal))
	}

	return append(ff,
		i.Name,
		min,
		max,
		strconv.Itoa(int(i.Status.PodDisruptionsAllowed)),
		strconv.Itoa(int(i.Status.CurrentHealthy)),
		strconv.Itoa(int(i.Status.DesiredHealthy)),
		strconv.Itoa(int(i.Status.ExpectedPods)),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// ExtFields returns extra info about the resource.
func (r *PodDisruptionBudget) ExtFields() Properties {
	return nil
}
