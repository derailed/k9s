package resource

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/k8s"
	v1beta1 "k8s.io/api/policy/v1beta1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// PodDisruptionBudget that can be displayed in a table and interacted with.
type PodDisruptionBudget struct {
	*Base

	instance *v1beta1.PodDisruptionBudget
}

// NewPDBList returns a new resource list.
func NewPDBList(c Connection, ns string) List {
	return NewList(
		ns,
		"pdb",
		NewPDB(c),
		AllVerbsAccess|DescribeAccess,
	)
}

// NewPDB instantiates a new PDB.
func NewPDB(c Connection) *PodDisruptionBudget {
	p := &PodDisruptionBudget{&Base{Connection: c, Resource: k8s.NewPodDisruptionBudget(c)}, nil}
	p.Factory = p

	return p
}

// New builds a new PDB instance from a k8s resource.
func (r *PodDisruptionBudget) New(i interface{}) (Columnar, error) {
	c := NewPDB(r.Connection)
	switch instance := i.(type) {
	case *v1beta1.PodDisruptionBudget:
		c.instance = instance
	case v1beta1.PodDisruptionBudget:
		c.instance = &instance
	case *interface{}:
		ptr := *i.(*interface{})
		pdbi, ok := ptr.(v1beta1.PodDisruptionBudget)
		if !ok {
			return nil, fmt.Errorf("Expecting PDB but got %T", ptr)
		}
		c.instance = &pdbi
	default:
		return nil, fmt.Errorf("Expecting PDB but got %T", instance)
	}
	c.path = c.namespacedName(c.instance.ObjectMeta)

	return c, nil
}

// Marshal resource to yaml.
func (r *PodDisruptionBudget) Marshal(path string) (string, error) {
	ns, n := Namespaced(path)
	i, err := r.Resource.Get(ns, n)
	if err != nil {
		return "", err
	}

	pdb, ok := i.(*v1beta1.PodDisruptionBudget)
	if !ok {
		return "", errors.New("Expecting a pdb resource")
	}
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

	return append(ff,
		i.Name,
		numbToStr(i.Spec.MinAvailable),
		numbToStr(i.Spec.MaxUnavailable),
		strconv.Itoa(int(i.Status.PodDisruptionsAllowed)),
		strconv.Itoa(int(i.Status.CurrentHealthy)),
		strconv.Itoa(int(i.Status.DesiredHealthy)),
		strconv.Itoa(int(i.Status.ExpectedPods)),
		toAge(i.ObjectMeta.CreationTimestamp),
	)
}

// Helpers...

func numbToStr(n *intstr.IntOrString) string {
	if n == nil {
		return NAValue
	}
	return strconv.Itoa(int(n.IntVal))
}
