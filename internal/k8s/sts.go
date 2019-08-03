package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StatefulSet manages a Kubernetes StatefulSet.
type StatefulSet struct {
	*base
	Connection
}

// NewStatefulSet instantiates a new StatefulSet.
func NewStatefulSet(c Connection) *StatefulSet {
	return &StatefulSet{&base{}, c}
}

// Get a StatefulSet.
func (s *StatefulSet) Get(ns, n string) (interface{}, error) {
	return s.DialOrDie().AppsV1().StatefulSets(ns).Get(n, metav1.GetOptions{})
}

// List all StatefulSets in a given namespace.
func (s *StatefulSet) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: s.labelSelector,
		FieldSelector: s.fieldSelector,
	}
	rr, err := s.DialOrDie().AppsV1().StatefulSets(ns).List(opts)
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
func (s *StatefulSet) Delete(ns, n string, cascade, force bool) error {
	p := metav1.DeletePropagationOrphan
	if cascade {
		p = metav1.DeletePropagationBackground
	}
	return s.DialOrDie().AppsV1().StatefulSets(ns).Delete(n, &metav1.DeleteOptions{
		PropagationPolicy: &p,
	})
}

// Scale a StatefulSet.
func (s *StatefulSet) Scale(ns, n string, replicas int32) error {
	scale, err := s.DialOrDie().Apps().StatefulSets(ns).GetScale(n, metav1.GetOptions{})
	if err != nil {
		return err
	}

	scale.Spec.Replicas = replicas
	_, err = s.DialOrDie().Apps().StatefulSets(ns).UpdateScale(n, scale)
	return err
}
