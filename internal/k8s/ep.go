package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Endpoints represents a Kubernetes Endpoints.
type Endpoints struct {
	*Resource
	Connection
}

// NewEndpoints returns a new Endpoints.
func NewEndpoints(c Connection, gvr GVR) *Endpoints {
	return &Endpoints{&Resource{gvr: gvr}, c}
}

// Get a Endpoint.
func (e *Endpoints) Get(ns, n string) (interface{}, error) {
	return e.DialOrDie().CoreV1().Endpoints(ns).Get(n, metav1.GetOptions{})
}

// List all Endpoints in a given namespace.
func (e *Endpoints) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: e.labelSelector,
		FieldSelector: e.fieldSelector,
	}
	rr, err := e.DialOrDie().CoreV1().Endpoints(ns).List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a Endpoint.
func (e *Endpoints) Delete(ns, n string, cascade, force bool) error {
	return e.DialOrDie().CoreV1().Endpoints(ns).Delete(n, nil)
}
