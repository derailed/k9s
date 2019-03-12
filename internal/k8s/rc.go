package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReplicationController represents a Kubernetes service.
type ReplicationController struct{}

// NewReplicationController returns a new ReplicationController.
func NewReplicationController() Res {
	return &ReplicationController{}
}

// Get a RC.
func (*ReplicationController) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().Core().ReplicationControllers(ns).Get(n, opts)
}

// List all RCs in a given namespace.
func (*ReplicationController) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().Core().ReplicationControllers(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a RC.
func (*ReplicationController) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().Core().ReplicationControllers(ns).Delete(n, &opts)
}
