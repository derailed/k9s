package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodDisruptionBudget represents a PodDisruptionBudget Kubernetes resource.
type PodDisruptionBudget struct {
	Connection
}

// NewPodDisruptionBudget returns a new PodDisruptionBudget.
func NewPodDisruptionBudget(c Connection) Cruder {
	return &PodDisruptionBudget{c}
}

// Get a pdb.
func (p *PodDisruptionBudget) Get(ns, n string) (interface{}, error) {
	return p.DialOrDie().PolicyV1beta1().PodDisruptionBudgets(ns).Get(n, metav1.GetOptions{})
}

// List all pdbs in a given namespace.
func (p *PodDisruptionBudget) List(ns string) (Collection, error) {
	rr, err := p.DialOrDie().PolicyV1beta1().PodDisruptionBudgets(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a pdb.
func (p *PodDisruptionBudget) Delete(ns, n string) error {
	return p.DialOrDie().PolicyV1beta1().PodDisruptionBudgets(ns).Delete(n, nil)
}
