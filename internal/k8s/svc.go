package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Service represents a Kubernetes Service.
type Service struct{}

// NewService returns a new Service.
func NewService() Res {
	return &Service{}
}

// Get a service.
func (*Service) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().CoreV1().Services(ns).Get(n, opts)
}

// List all Services in a given namespace.
func (*Service) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().CoreV1().Services(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a Service.
func (*Service) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().CoreV1().Services(ns).Delete(n, &opts)
}
