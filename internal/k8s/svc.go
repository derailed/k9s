package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Service represents a Kubernetes Service.
type Service struct {
	*base
	Connection
}

// NewService returns a new Service.
func NewService(c Connection) *Service {
	return &Service{&base{}, c}
}

// Get a service.
func (s *Service) Get(ns, n string) (interface{}, error) {
	return s.DialOrDie().CoreV1().Services(ns).Get(n, metav1.GetOptions{})
}

// List all Services in a given namespace.
func (s *Service) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: s.labelSelector,
		FieldSelector: s.fieldSelector,
	}
	rr, err := s.DialOrDie().CoreV1().Services(ns).List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a Service.
func (s *Service) Delete(ns, n string) error {
	return s.DialOrDie().CoreV1().Services(ns).Delete(n, nil)
}
