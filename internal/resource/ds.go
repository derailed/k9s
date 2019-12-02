package resource

// import (
// 	"context"
// 	"errors"
// 	"fmt"
// 	"strconv"

// 	"github.com/derailed/k9s/internal"
// 	"github.com/derailed/k9s/internal/k8s"
// 	"github.com/derailed/k9s/internal/watch"
// 	appsv1 "k8s.io/api/apps/v1"
// 	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
// 	"k8s.io/apimachinery/pkg/labels"
// 	"k8s.io/apimachinery/pkg/runtime"
// )

// // Compile time checks to ensure type satisfies interface
// var _ Restartable = (*DaemonSet)(nil)

// // DaemonSet tracks a kubernetes resource.
// type DaemonSet struct {
// 	*Base
// 	instance *appsv1.DaemonSet
// }

// // NewDaemonSetList returns a new resource list.
// func NewDaemonSetList(c Connection, ns string) List {
// 	return NewList(
// 		ns,
// 		"ds",
// 		NewDaemonSet(c),
// 		AllVerbsAccess|DescribeAccess,
// 	)
// }

// // NewDaemonSet instantiates a new DaemonSet.
// func NewDaemonSet(c Connection) *DaemonSet {
// 	ds := &DaemonSet{&Base{Connection: c, Resource: k8s.NewDaemonSet(c)}, nil}
// 	ds.Factory = ds

// 	return ds
// }

// // New builds a new DaemonSet instance from a k8s resource.
// func (r *DaemonSet) New(i interface{}) (Columnar, error) {
// 	c := NewDaemonSet(r.Connection)
// 	switch instance := i.(type) {
// 	case *appsv1.DaemonSet:
// 		c.instance = instance
// 	case appsv1.DaemonSet:
// 		c.instance = &instance
// 	default:
// 		return nil, fmt.Errorf("Expecting DaemonSet but got %T", instance)
// 	}
// 	c.path = c.namespacedName(c.instance.ObjectMeta)

// 	return c, nil
// }

// // Marshal resource to yaml.
// func (r *DaemonSet) Marshal(path string) (string, error) {
// 	ns, n := Namespaced(path)
// 	i, err := r.Resource.Get(ns, n)
// 	if err != nil {
// 		return "", err
// 	}

// 	ds, ok := i.(*appsv1.DaemonSet)
// 	if !ok {
// 		return "", errors.New("expecting ds resource")
// 	}
// 	ds.TypeMeta.APIVersion = "apps/v1"
// 	ds.TypeMeta.Kind = "DaemonSet"

// 	return r.marshalObject(ds)
// }

// // Logs tail logs for all pods represented by this DaemonSet.
// func (r *DaemonSet) Logs(ctx context.Context, c chan<- string, opts LogOptions) error {
// 	f, ok := ctx.Value(internal.KeyFactory).(*watch.Factory)
// 	if !ok {
// 		return errors.New("no factory in context for pod logs")
// 	}

// 	o, err := f.Get(opts.Namespace, "apps/v1/daemonsets", opts.Name, labels.Everything())
// 	if err != nil {
// 		return err
// 	}

// 	var ds appsv1.DaemonSet
// 	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
// 	if err != nil {
// 		return errors.New("expecting daemonset resource")
// 	}

// 	if ds.Spec.Selector == nil || len(ds.Spec.Selector.MatchLabels) == 0 {
// 		return fmt.Errorf("No valid selector found on daemonset %s", opts.FQN())
// 	}

// 	return r.podLogs(ctx, c, ds.Spec.Selector.MatchLabels, opts)
// }

// // Header return resource header.
// func (*DaemonSet) Header(ns string) Row {
// 	hh := Row{}
// 	if ns == AllNamespaces {
// 		hh = append(hh, "NAMESPACE")
// 	}
// 	hh = append(hh, "NAME", "DESIRED", "CURRENT", "READY", "UP-TO-DATE")
// 	hh = append(hh, "AVAILABLE", "NODE_SELECTOR", "AGE")

// 	return hh
// }

// // Fields retrieves displayable fields.
// func (r *DaemonSet) Fields(ns string) Row {
// 	ff := make([]string, 0, len(r.Header(ns)))

// 	i := r.instance
// 	if ns == AllNamespaces {
// 		ff = append(ff, i.Namespace)
// 	}

// 	return append(ff,
// 		i.Name,
// 		strconv.Itoa(int(i.Status.DesiredNumberScheduled)),
// 		strconv.Itoa(int(i.Status.CurrentNumberScheduled)),
// 		strconv.Itoa(int(i.Status.NumberReady)),
// 		strconv.Itoa(int(i.Status.UpdatedNumberScheduled)),
// 		strconv.Itoa(int(i.Status.NumberAvailable)),
// 		mapToStr(i.Spec.Template.Spec.NodeSelector),
// 		toAge(i.ObjectMeta.CreationTimestamp),
// 	)
// }

// // Restart the rollout of the specified resource.
// func (r *DaemonSet) Restart(ns, n string) error {
// 	return r.Resource.(Restartable).Restart(ns, n)
// }
