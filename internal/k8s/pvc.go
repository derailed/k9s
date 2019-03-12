package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PVC represents a Kubernetes service.
type PVC struct{}

// NewPVC returns a new PVC.
func NewPVC() Res {
	return &PVC{}
}

// Get a PVC.
func (*PVC) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().CoreV1().PersistentVolumeClaims(ns).Get(n, opts)
}

// List all PVCs in a given namespace.
func (*PVC) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().CoreV1().PersistentVolumeClaims(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a PVC.
func (*PVC) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().CoreV1().PersistentVolumeClaims(ns).Delete(n, &opts)
}
