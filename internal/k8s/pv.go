package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PeristentVolume represents a Kubernetes PersistentVolume.
type PeristentVolume struct {
	*base
	Connection
}

// NewPersistentVolume returns a new PeristentVolume.
func NewPersistentVolume(c Connection) *PeristentVolume {
	return &PeristentVolume{&base{}, c}
}

// Get a PeristentVolume.
func (p *PeristentVolume) Get(_, n string) (interface{}, error) {
	return p.DialOrDie().CoreV1().PersistentVolumes().Get(n, metav1.GetOptions{})
}

// List all PeristentVolumes in a given namespace.
func (p *PeristentVolume) List(_ string) (Collection, error) {
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

// Delete a PeristentVolume.
func (p *PeristentVolume) Delete(_, n string) error {
	return p.DialOrDie().CoreV1().PersistentVolumes().Delete(n, nil)
}
