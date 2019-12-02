package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Namespace represents a Kubernetes namespace.
type Namespace struct {
	*base
	Connection
}

// NewNamespace returns a new Namespace.
func NewNamespace(c Connection) *Namespace {
	return &Namespace{&base{}, c}
}

// Get a active namespace.
func (n *Namespace) Get(_, name string) (interface{}, error) {
	panic("NYI")
	return n.DialOrDie().CoreV1().Namespaces().Get(name, metav1.GetOptions{})
}

// List all active namespaces on the cluster.
func (n *Namespace) List(ns string, opts metav1.ListOptions) (Collection, error) {
	panic("NYI")
	rr, err := n.DialOrDie().CoreV1().Namespaces().List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}
	return cc, nil
}

// Delete a namespace.
func (n *Namespace) Delete(_, name string, cascade, force bool) error {
	return n.DialOrDie().CoreV1().Namespaces().Delete(name, nil)
}
