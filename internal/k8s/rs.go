package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReplicaSet represents a Kubernetes ReplicaSet.
type ReplicaSet struct {
	*base
	Connection
}

// NewReplicaSet returns a new ReplicaSet.
func NewReplicaSet(c Connection) *ReplicaSet {
	return &ReplicaSet{&base{}, c}
}

// Get a ReplicaSet.
func (r *ReplicaSet) Get(ns, n string) (interface{}, error) {
	return r.DialOrDie().AppsV1().ReplicaSets(ns).Get(n, metav1.GetOptions{})
}

// List all ReplicaSets in a given namespace.
func (r *ReplicaSet) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: r.labelSelector,
		FieldSelector: r.fieldSelector,
	}
	rr, err := r.DialOrDie().AppsV1().ReplicaSets(ns).List(opts)
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
func (r *ReplicaSet) Delete(ns, n string, cascade, force bool) error {
	return r.DialOrDie().AppsV1().ReplicaSets(ns).Delete(n, nil)
}
