package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PersistentVolume represents a Kubernetes PersistentVolume.
type PersistentVolume struct {
	*base
	Connection
}

// NewPersistentVolume returns a new PersistentVolume.
func NewPersistentVolume(c Connection) *PersistentVolume {
	return &PersistentVolume{&base{}, c}
}

// Get a PersistentVolume.
func (p *PersistentVolume) Get(_, n string) (interface{}, error) {
	return p.DialOrDie().CoreV1().PersistentVolumes().Get(n, metav1.GetOptions{})
}

// List all PersistentVolumes in a given namespace.
func (p *PersistentVolume) List(_ string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: p.labelSelector,
		FieldSelector: p.fieldSelector,
	}
	rr, err := p.DialOrDie().CoreV1().PersistentVolumes().List(opts)
	if err != nil {
		return nil, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a PersistentVolume.
func (p *PersistentVolume) Delete(_, n string, cascade, force bool) error {
	return p.DialOrDie().CoreV1().PersistentVolumes().Delete(n, nil)
}
