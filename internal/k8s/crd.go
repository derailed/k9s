package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CRD represents a Kubernetes CRD
type CRD struct{}

// NewCRD returns a new CRD.
func NewCRD() Res {
	return &CRD{}
}

// Get a CRD.
func (*CRD) Get(_, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.nsDialOrDie().Get(n, opts)
}

// List all CRDs in a given namespace.
func (*CRD) List(string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.nsDialOrDie().List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a CRD.
func (*CRD) Delete(_, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.nsDialOrDie().Delete(n, &opts)
}
