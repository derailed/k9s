package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Endpoints represents a Kubernetes Endpoints
type Endpoints struct{}

// NewEndpoints returns a new Endpoints.
func NewEndpoints() Res {
	return &Endpoints{}
}

// Get a service.
func (*Endpoints) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().CoreV1().Endpoints(ns).Get(n, opts)
}

// List all Endpointss in a given namespace
func (*Endpoints) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().CoreV1().Endpoints(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a Endpoints
func (*Endpoints) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().CoreV1().Endpoints(ns).Delete(n, &opts)
}
