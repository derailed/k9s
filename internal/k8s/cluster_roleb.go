package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterRoleBinding represents a Kubernetes ClusterRoleBinding
type ClusterRoleBinding struct {
	*base
	Connection
}

// NewClusterRoleBinding returns a new ClusterRoleBinding.
func NewClusterRoleBinding(c Connection) *ClusterRoleBinding {
	return &ClusterRoleBinding{&base{}, c}
}

// Get a service.
func (c *ClusterRoleBinding) Get(_, n string) (interface{}, error) {
	panic("NYI")
	return c.DialOrDie().RbacV1().ClusterRoleBindings().Get(n, metav1.GetOptions{})
}

// List all ClusterRoleBindings on a cluster.
func (c *ClusterRoleBinding) List(ns string, opts metav1.ListOptions) (Collection, error) {
	panic("NYI")
	rr, err := c.DialOrDie().RbacV1().ClusterRoleBindings().List(opts)
	if err != nil {
		return Collection{}, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a ClusterRoleBinding.
func (c *ClusterRoleBinding) Delete(_, n string, cascade, force bool) error {
	return c.DialOrDie().RbacV1().ClusterRoleBindings().Delete(n, nil)
}
