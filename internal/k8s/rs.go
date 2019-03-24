package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReplicaSet represents a Kubernetes ReplicaSet.
type ReplicaSet struct {
	Connection
}

// NewReplicaSet returns a new ReplicaSet.
func NewReplicaSet(c Connection) Cruder {
	return &ReplicaSet{c}
}

// Get a ReplicaSet.
func (r *ReplicaSet) Get(ns, n string) (interface{}, error) {
	return r.DialOrDie().Apps().ReplicaSets(ns).Get(n, metav1.GetOptions{})
}

// List all ReplicaSets in a given namespace.
func (r *ReplicaSet) List(ns string) (Collection, error) {
	rr, err := r.DialOrDie().Apps().ReplicaSets(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a ReplicaSet.
func (r *ReplicaSet) Delete(ns, n string) error {
	return r.DialOrDie().Apps().ReplicaSets(ns).Delete(n, nil)
}
