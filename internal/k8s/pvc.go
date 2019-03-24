package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PVC represents a Kubernetes service.
type PVC struct {
	Connection
}

// NewPVC returns a new PVC.
func NewPVC(c Connection) Cruder {
	return &PVC{c}
}

// Get a PVC.
func (p *PVC) Get(ns, n string) (interface{}, error) {
	return p.DialOrDie().CoreV1().PersistentVolumeClaims(ns).Get(n, metav1.GetOptions{})
}

// List all PVCs in a given namespace.
func (p *PVC) List(ns string) (Collection, error) {
	rr, err := p.DialOrDie().CoreV1().PersistentVolumeClaims(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a PVC.
func (p *PVC) Delete(ns, n string) error {
	return p.DialOrDie().CoreV1().PersistentVolumeClaims(ns).Delete(n, nil)
}
