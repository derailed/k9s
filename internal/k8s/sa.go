package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceAccount manages a Kubernetes ServiceAccount.
type ServiceAccount struct {
	Connection
}

// NewServiceAccount instantiates a new ServiceAccount.
func NewServiceAccount(c Connection) Cruder {
	return &ServiceAccount{c}
}

// Get a ServiceAccount.
func (s *ServiceAccount) Get(ns, n string) (interface{}, error) {
	return s.DialOrDie().CoreV1().ServiceAccounts(ns).Get(n, metav1.GetOptions{})
}

// List all ServiceAccounts in a given namespace.
func (s *ServiceAccount) List(ns string) (Collection, error) {
	rr, err := s.DialOrDie().CoreV1().ServiceAccounts(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}
	return cc, nil

}

// Delete a ServiceAccount.
func (s *ServiceAccount) Delete(ns, n string) error {
	return s.DialOrDie().CoreV1().ServiceAccounts(ns).Delete(n, nil)
}
