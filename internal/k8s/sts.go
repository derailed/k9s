package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatefulSet manages a Kubernetes StatefulSet.
type StatefulSet struct{}

// NewStatefulSet instantiates a new StatefulSet.
func NewStatefulSet() Res {
	return &StatefulSet{}
}

// Get a StatefulSet.
func (*StatefulSet) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	o, err := conn.dialOrDie().AppsV1().StatefulSets(ns).Get(n, opts)
	if err != nil {
		return o, err
	}
	return o, nil
}

// List all StatefulSets in a given namespace.
func (*StatefulSet) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().AppsV1().StatefulSets(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}
	return cc, nil
}

// Delete a StatefulSet.
func (*StatefulSet) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().AppsV1().StatefulSets(ns).Delete(n, &opts)
}
