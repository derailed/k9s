package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatefulSet manages a Kubernetes StatefulSet.
type StatefulSet struct {
	Connection
}

// NewStatefulSet instantiates a new StatefulSet.
func NewStatefulSet(c Connection) Cruder {
	return &StatefulSet{c}
}

// Get a StatefulSet.
func (s *StatefulSet) Get(ns, n string) (interface{}, error) {
	return s.DialOrDie().AppsV1().StatefulSets(ns).Get(n, metav1.GetOptions{})
}

// List all StatefulSets in a given namespace.
func (s *StatefulSet) List(ns string) (Collection, error) {
	rr, err := s.DialOrDie().AppsV1().StatefulSets(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a StatefulSet.
func (s *StatefulSet) Delete(ns, n string) error {
	return s.DialOrDie().AppsV1().StatefulSets(ns).Delete(n, nil)
}
