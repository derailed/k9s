package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReplicationController represents a Kubernetes service.
type ReplicationController struct {
	Connection
}

// NewReplicationController returns a new ReplicationController.
func NewReplicationController(c Connection) Cruder {
	return &ReplicationController{c}
}

// Get a RC.
func (r *ReplicationController) Get(ns, n string) (interface{}, error) {
	return r.DialOrDie().Core().ReplicationControllers(ns).Get(n, metav1.GetOptions{})
}

// List all RCs in a given namespace.
func (r *ReplicationController) List(ns string) (Collection, error) {
	rr, err := r.DialOrDie().Core().ReplicationControllers(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a RC.
func (r *ReplicationController) Delete(ns, n string) error {
	return r.DialOrDie().Core().ReplicationControllers(ns).Delete(n, nil)
}
