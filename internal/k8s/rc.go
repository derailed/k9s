package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReplicationController represents a Kubernetes ReplicationController.
type ReplicationController struct {
	*base
	Connection
}

// NewReplicationController returns a new ReplicationController.
func NewReplicationController(c Connection) *ReplicationController {
	return &ReplicationController{&base{}, c}
}

// Get a RC.
func (r *ReplicationController) Get(ns, n string) (interface{}, error) {
	return r.DialOrDie().Core().ReplicationControllers(ns).Get(n, metav1.GetOptions{})
}

// List all RCs in a given namespace.
func (r *ReplicationController) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: r.labelSelector,
		FieldSelector: r.fieldSelector,
	}
	rr, err := r.DialOrDie().Core().ReplicationControllers(ns).List(opts)
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
func (r *ReplicationController) Delete(ns, n string, cascade, force bool) error {
	p := metav1.DeletePropagationOrphan
	if cascade {
		p = metav1.DeletePropagationBackground
	}
	return r.DialOrDie().Core().ReplicationControllers(ns).Delete(n, &metav1.DeleteOptions{
		PropagationPolicy: &p,
	})
}

// Scale a ReplicationController.
func (r *ReplicationController) Scale(ns, n string, replicas int32) error {
	scale, err := r.DialOrDie().Core().ReplicationControllers(ns).GetScale(n, metav1.GetOptions{})
	if err != nil {
		return err
	}

	scale.Spec.Replicas = replicas
	_, err = r.DialOrDie().Core().ReplicationControllers(ns).UpdateScale(n, scale)
	return err
}
