package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodDisruptionBudget represents a PodDisruptionBudget Kubernetes resource.
type PodDisruptionBudget struct{}

// NewPodDisruptionBudget returns a new PodDisruptionBudget.
func NewPodDisruptionBudget() Res {
	return &PodDisruptionBudget{}
}

// Get a pdb.
func (*PodDisruptionBudget) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().PolicyV1beta1().PodDisruptionBudgets(ns).Get(n, opts)
}

// List all pdbs in a given namespace.
func (*PodDisruptionBudget) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().PolicyV1beta1().PodDisruptionBudgets(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a pdb.
func (*PodDisruptionBudget) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().PolicyV1beta1().PodDisruptionBudgets(ns).Delete(n, &opts)
}
