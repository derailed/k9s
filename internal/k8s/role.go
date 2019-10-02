package k8s

import (
	// rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Role represents a Kubernetes Role.
type Role struct {
	*Resource
	Connection
}

// NewRole returns a new Role.
func NewRole(c Connection, gvr GVR) *Role {
	return &Role{&Resource{gvr: gvr}, c}
}

// Get a Role.
func (r *Role) Get(ns, n string) (interface{}, error) {
	return r.DialOrDie().RbacV1().Roles(ns).Get(n, metav1.GetOptions{})
}

// List all Roles in a given namespace.
func (r *Role) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: r.labelSelector,
		FieldSelector: r.fieldSelector,
	}
	rr, err := r.DialOrDie().RbacV1().Roles(ns).List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a Role.
func (r *Role) Delete(ns, n string, cascade, force bool) error {
	return r.DialOrDie().RbacV1().Roles(ns).Delete(n, nil)
}
