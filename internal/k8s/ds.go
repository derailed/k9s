package k8s

// BOZO!!
// import (
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// 	"k8s.io/apimachinery/pkg/types"
// 	"k8s.io/kubectl/pkg/polymorphichelpers"
// )

// // DaemonSet represents a Kubernetes DaemonSet
// type DaemonSet struct {
// 	*base
// 	Connection
// }

// // NewDaemonSet returns a new DaemonSet.
// func NewDaemonSet(c Connection) *DaemonSet {
// 	return &DaemonSet{&base{}, c}
// }

// // Get a DaemonSet.
// func (d *DaemonSet) Get(ns, n string) (interface{}, error) {
// 	panic("NYI")
// 	return d.DialOrDie().AppsV1().DaemonSets(ns).Get(n, metav1.GetOptions{})
// }

// // List all DaemonSets in a given namespace.
// func (d *DaemonSet) List(ns string, opts metav1.ListOptions) (Collection, error) {
// 	panic("NYI")
// 	rr, err := d.DialOrDie().AppsV1().DaemonSets(ns).List(opts)
// 	if err != nil {
// 		return nil, err
// 	}
// 	cc := make(Collection, len(rr.Items))
// 	for i, r := range rr.Items {
// 		cc[i] = r
// 	}

// 	return cc, nil
// }

// // Delete a DaemonSet.
// func (d *DaemonSet) Delete(ns, n string, cascade, force bool) error {
// 	p := metav1.DeletePropagationOrphan
// 	if cascade {
// 		p = metav1.DeletePropagationBackground
// 	}
// 	return d.DialOrDie().AppsV1().DaemonSets(ns).Delete(n, &metav1.DeleteOptions{
// 		PropagationPolicy: &p,
// 	})
// }

// // Restart a DaemonSet rollout.
// func (d *DaemonSet) Restart(f *watch.Factory, ns, n string) error {
// 	o, err := f.Get(ns, "apps/v1/deamonsets", n, labels.Everything())
// 	if err != nil {
// 		return err
// 	}

// 	var ds appsv1.DaemonSet
// 	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
// 	if err != nil {
// 		return err
// 	}

// 	update, err := polymorphichelpers.ObjectRestarterFn(ds)
// 	if err != nil {
// 		return err
// 	}

// 	_, err = f.Client().DialOrDie().AppsV1().DaemonSets(ns).Patch(ds.Name, types.StrategicMergePatchType, update)
// 	return err
// }
