package k8s

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// RoleBinding represents a Kubernetes service
type RoleBinding struct{}

// NewRoleBinding returns a new RoleBinding.
func NewRoleBinding() Res {
	return &RoleBinding{}
}

// Get a service.
func (*RoleBinding) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().RbacV1().RoleBindings(ns).Get(n, opts)
}

// List all services in a given namespace
func (*RoleBinding) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().RbacV1().RoleBindings(ns).List(opts)
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
func (*RoleBinding) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().RbacV1().RoleBindings(ns).Delete(n, &opts)
}
