package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReplicaSet represents a Kubernetes ReplicaSet.
type ReplicaSet struct{}

// NewReplicaSet returns a new ReplicaSet.
func NewReplicaSet() Res {
	return &ReplicaSet{}
}

// Get a ReplicaSet.
func (*ReplicaSet) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().Apps().ReplicaSets(ns).Get(n, opts)
}

// List all ReplicaSets in a given namespace.
func (*ReplicaSet) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().Apps().ReplicaSets(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a ReplicaSet.
func (*ReplicaSet) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().Apps().ReplicaSets(ns).Delete(n, &opts)
}
