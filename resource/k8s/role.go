package k8s

import (
	// rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Role represents a Kubernetes service
type Role struct{}

// NewRole returns a new Role.
func NewRole() Res {
	return &Role{}
}

// Get a service.
func (*Role) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().RbacV1().Roles(ns).Get(n, opts)
}

// List all services in a given namespace
func (*Role) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().RbacV1().Roles(ns).List(opts)
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
func (*Role) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().RbacV1().Roles(ns).Delete(n, &opts)
}
