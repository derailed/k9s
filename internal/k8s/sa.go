package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServiceAccount manages a Kubernetes ServiceAccount.
type ServiceAccount struct{}

// NewServiceAccount instantiates a new ServiceAccount.
func NewServiceAccount() Res {
	return &ServiceAccount{}
}

// Get a ServiceAccount
func (*ServiceAccount) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	o, err := conn.dialOrDie().CoreV1().ServiceAccounts(ns).Get(n, opts)
	if err != nil {
		return o, err
	}
	return o, nil
}

// List all ServiceAccounts in a given namespace
func (*ServiceAccount) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().CoreV1().ServiceAccounts(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}
	return cc, nil

}

// Delete a ServiceAccount
func (*ServiceAccount) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().CoreV1().ServiceAccounts(ns).Delete(n, &opts)
}
