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
func NewPDBList(c k8s.Connection, ns string) List {
	return NewList(
		ns,
		"pdb",
		NewPDB(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewPDB instantiates a new PDB.
func NewPDB(c k8s.Connection) *PodDisruptionBudget {
	p := &PodDisruptionBudget{&Base{Connection: c, Resource: k8s.NewPodDisruptionBudget(c)}, nil}
	p.Factory = p

	return p
}

// New builds a new PDB instance from a k8s resource.
func (r *PodDisruptionBudget) New(i interface{}) Columnar {
	c := NewPDB(r.Connection)
	switch instance := i.(type) {
	case *v1beta1.PodDisruptionBudget:
		c.instance = instance
	case v1beta1.PodDisruptionBudget:
		c.instance = &instance
	case *interface{}:
		ptr := *i.(*interface{})
		pdbi := ptr.(v1beta1.PodDisruptionBudget)
		c.instance = &pdbi
	default:
		log.Fatal().Msgf("unknown PDB type %#v", i)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c
}

// Marshal resource to yaml.
func (r *PodDisruptionBudget) Marshal(path string) (string, error) {
	ns, n := namespaced(path)
	i, err := r.Resource.Get(ns, n)
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
