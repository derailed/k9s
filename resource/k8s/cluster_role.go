package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterRole represents a Kubernetes ClusterRole
type ClusterRole struct{}

// NewClusterRole returns a new ClusterRole.
func NewClusterRole() Res {
	return &ClusterRole{}
}

// Get a service.
func (*ClusterRole) Get(_, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().RbacV1().ClusterRoles().Get(n, opts)
}

// List all ClusterRoles in a given namespace
func (*ClusterRole) List(_ string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().RbacV1().ClusterRoles().List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a ClusterRole
func (*ClusterRole) Delete(_, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().RbacV1().ClusterRoles().Delete(n, &opts)
}
