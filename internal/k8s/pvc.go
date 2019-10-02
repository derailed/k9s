package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistentVolumeClaim represents a Kubernetes PersistentVolumeClaim.
type PersistentVolumeClaim struct {
	*Resource
	Connection
}

// NewPersistentVolumeClaim returns a new PersistentVolumeClaim.
func NewPersistentVolumeClaim(c Connection, gvr GVR) *PersistentVolumeClaim {
	return &PersistentVolumeClaim{&Resource{gvr: gvr}, c}
}

// Get a PersistentVolumeClaim.
func (p *PersistentVolumeClaim) Get(ns, n string) (interface{}, error) {
	return p.DialOrDie().CoreV1().PersistentVolumeClaims(ns).Get(n, metav1.GetOptions{})
}

// List all PersistentVolumeClaims in a given namespace.
func (p *PersistentVolumeClaim) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: p.labelSelector,
		FieldSelector: p.fieldSelector,
	}
	rr, err := p.DialOrDie().CoreV1().PersistentVolumeClaims(ns).List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a PersistentVolumeClaim.
func (p *PersistentVolumeClaim) Delete(ns, n string, cascade, force bool) error {
	return p.DialOrDie().CoreV1().PersistentVolumeClaims(ns).Delete(n, nil)
}
