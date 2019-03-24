package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CRD represents a Kubernetes CRD
type CRD struct {
	Connection
}

// NewCRD returns a new CRD.
func NewCRD(c Connection) Cruder {
	return &CRD{c}
}

// Get a CRD.
func (c *CRD) Get(_, n string) (interface{}, error) {
	return c.NSDialOrDie().Get(n, metav1.GetOptions{})
}

// List all CRDs in a given namespace.
func (c *CRD) List(string) (Collection, error) {
	rr, err := c.NSDialOrDie().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a CRD.
func (c *CRD) Delete(_, n string) error {
	return c.NSDialOrDie().Delete(n, nil)
}
