package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterRole represents a Kubernetes ClusterRole
type ClusterRole struct {
	Connection
}

// NewClusterRole returns a new ClusterRole.
func NewClusterRole(c Connection) Cruder {
	return &ClusterRole{c}
}

// Get a cluster role.
func (c *ClusterRole) Get(_, n string) (interface{}, error) {
	return c.DialOrDie().RbacV1().ClusterRoles().Get(n, metav1.GetOptions{})
}

// List all ClusterRoles on a cluster.
func (c *ClusterRole) List(_ string) (Collection, error) {
	rr, err := c.DialOrDie().RbacV1().ClusterRoles().List(metav1.ListOptions{})
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
func (c *ClusterRole) Delete(_, n string) error {
	return c.DialOrDie().RbacV1().ClusterRoles().Delete(n, nil)
}
