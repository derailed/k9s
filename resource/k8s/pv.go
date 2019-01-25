package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PV represents a Kubernetes service
type PV struct{}

// NewPV returns a new PV.
func NewPV() Res {
	return &PV{}
}

// Get a service.
func (*PV) Get(_, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().CoreV1().PersistentVolumes().Get(n, opts)
}

// List all services in a given namespace
func (*PV) List(_ string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().CoreV1().PersistentVolumes().List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a service
func (*PV) Delete(_, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().CoreV1().PersistentVolumes().Delete(n, &opts)
}
