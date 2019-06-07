package k8s

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// RoleBinding represents a Kubernetes RoleBinding.
type RoleBinding struct {
	*base
	Connection
}

// NewRoleBinding returns a new RoleBinding.
func NewRoleBinding(c Connection) *RoleBinding {
	return &RoleBinding{&base{}, c}
}

// Get a RoleBinding.
func (r *RoleBinding) Get(ns, n string) (interface{}, error) {
	return r.DialOrDie().RbacV1().RoleBindings(ns).Get(n, metav1.GetOptions{})
}

// List all RoleBindings in a given namespace.
func (r *RoleBinding) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: r.labelSelector,
		FieldSelector: r.fieldSelector,
	}
	rr, err := r.DialOrDie().RbacV1().RoleBindings(ns).List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a RoleBinding.
func (r *RoleBinding) Delete(ns, n string, cascade, force bool) error {
	return r.DialOrDie().RbacV1().RoleBindings(ns).Delete(n, nil)
}
