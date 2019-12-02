package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterRole represents a Kubernetes ClusterRole
type ClusterRole struct {
	*base
	Connection
}

// NewClusterRole returns a new ClusterRole.
func NewClusterRole(c Connection) *ClusterRole {
	return &ClusterRole{&base{}, c}
}

// Get a cluster role.
func (c *ClusterRole) Get(_, n string) (interface{}, error) {
	panic("NYI")
	return c.DialOrDie().RbacV1().ClusterRoles().Get(n, metav1.GetOptions{})
}

// List all ClusterRoles on a cluster.
func (c *ClusterRole) List(ns string, opts metav1.ListOptions) (Collection, error) {
	panic("NYI")
	rr, err := c.DialOrDie().RbacV1().ClusterRoles().List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a ClusterRole.
func (c *ClusterRole) Delete(_, n string, cascade, force bool) error {
	return c.DialOrDie().RbacV1().ClusterRoles().Delete(n, nil)
}
