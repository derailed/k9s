package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterRoleBinding represents a Kubernetes ClusterRoleBinding
type ClusterRoleBinding struct{}

// NewClusterRoleBinding returns a new ClusterRoleBinding.
func NewClusterRoleBinding() Res {
	return &ClusterRoleBinding{}
}

// Get a service.
func (*ClusterRoleBinding) Get(_, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().RbacV1().ClusterRoleBindings().Get(n, opts)
}

// List all ClusterRoleBindings on a cluster.
func (*ClusterRoleBinding) List(_ string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().RbacV1().ClusterRoleBindings().List(opts)
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
func (*ClusterRoleBinding) Delete(_, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().RbacV1().ClusterRoleBindings().Delete(n, &opts)
}
