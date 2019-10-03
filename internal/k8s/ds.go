package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaemonSet represents a Kubernetes DaemonSet
type DaemonSet struct {
	*base
	Connection
}

// NewDaemonSet returns a new DaemonSet.
func NewDaemonSet(c Connection) *DaemonSet {
	return &DaemonSet{&base{}, c}
}

// Get a DaemonSet.
func (d *DaemonSet) Get(ns, n string) (interface{}, error) {
	return d.DialOrDie().AppsV1().DaemonSets(ns).Get(n, metav1.GetOptions{})
}

// List all DaemonSets in a given namespace.
func (d *DaemonSet) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: d.labelSelector,
		FieldSelector: d.fieldSelector,
	}
	rr, err := d.DialOrDie().AppsV1().DaemonSets(ns).List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a DaemonSet.
func (d *DaemonSet) Delete(ns, n string, cascade, force bool) error {
	p := metav1.DeletePropagationOrphan
	if cascade {
		p = metav1.DeletePropagationBackground
	}
	return d.DialOrDie().AppsV1().DaemonSets(ns).Delete(n, &metav1.DeleteOptions{
		PropagationPolicy: &p,
	})
}
