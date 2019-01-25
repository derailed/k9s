package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Node represents a Kubernetes service
type Node struct{}

// NewNode returns a new Node.
func NewNode() Res {
	return &Node{}
}

// Get a service.
func (*Node) Get(_, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().CoreV1().Nodes().Get(n, opts)
}

// List all services in a given namespace
func (*Node) List(_ string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().CoreV1().Nodes().List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a service
func (*Node) Delete(_, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().CoreV1().Nodes().Delete(n, &opts)
}
