package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PV represents a Kubernetes PersistentVolume.
type PV struct {
	Connection
}

// NewPV returns a new PV.
func NewPV(c Connection) Cruder {
	return &PV{c}
}

// Get a PV.
func (p *PV) Get(_, n string) (interface{}, error) {
	return p.DialOrDie().CoreV1().PersistentVolumes().Get(n, metav1.GetOptions{})
}

// List all PVs in a given namespace.
func (p *PV) List(_ string) (Collection, error) {
	rr, err := p.DialOrDie().CoreV1().PersistentVolumes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a PV.
func (p *PV) Delete(_, n string) error {
	return p.DialOrDie().CoreV1().PersistentVolumes().Delete(n, nil)
}
