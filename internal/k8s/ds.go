package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/polymorphichelpers"
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
func (d *DaemonSet) List(ns string, opts metav1.ListOptions) (Collection, error) {
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

// Restart a DaemonSet rollout.
func (d *DaemonSet) Restart(ns, n string) error {

	ds, err := d.DialOrDie().AppsV1().DaemonSets(ns).Get(n, metav1.GetOptions{})
	if err != nil {
		return err
	}
	update, err := polymorphichelpers.ObjectRestarterFn(ds)
	if err != nil {
		return err
	}

	_, err = d.DialOrDie().AppsV1().DaemonSets(ns).Patch(ds.Name, types.StrategicMergePatchType, update)
	return err
}
