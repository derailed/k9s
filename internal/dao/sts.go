package dao

import (
	"context"
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubectl/pkg/polymorphichelpers"
)

var (
	_ Accessor    = (*StatefulSet)(nil)
	_ Nuker       = (*StatefulSet)(nil)
	_ Loggable    = (*StatefulSet)(nil)
	_ Restartable = (*StatefulSet)(nil)
	_ Scalable    = (*StatefulSet)(nil)
)

// StatefulSet represents a K8s sts.
type StatefulSet struct {
	Resource
}

// Scale a StatefulSet.
func (s *StatefulSet) Scale(path string, replicas int32) error {
	ns, n := client.Namespaced(path)
	auth, err := s.Client().CanI(ns, "apps/v1/statefulsets:scale", []string{"get", "update"})
	if !auth || err != nil {
		return err
	}
	scale, err := s.Client().DialOrDie().AppsV1().StatefulSets(ns).GetScale(n, metav1.GetOptions{})
	if err != nil {
		return err
	}
	scale.Spec.Replicas = replicas
	_, err = s.Client().DialOrDie().AppsV1().StatefulSets(ns).UpdateScale(n, scale)

	return err
}

// Restart a StatefulSet rollout.
func (s *StatefulSet) Restart(path string) error {
	o, err := s.Factory.Get(s.gvr.String(), path, true, labels.Everything())
	if err != nil {
		return err
	}
	var ds appsv1.StatefulSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
	if err != nil {
		return err
	}

	ns, _ := client.Namespaced(path)
	auth, err := s.Client().CanI(ns, "apps/v1/statefulsets", []string{"patch"})
	if !auth || err != nil {
		return err
	}
	update, err := polymorphichelpers.ObjectRestarterFn(&ds)
	if err != nil {
		return err
	}
	_, err = s.Client().DialOrDie().AppsV1().StatefulSets(ds.Namespace).Patch(ds.Name, types.StrategicMergePatchType, update)

	return err
}

// TailLogs tail logs for all pods represented by this StatefulSet.
func (s *StatefulSet) TailLogs(ctx context.Context, c chan<- string, opts LogOptions) error {
	o, err := s.Factory.Get(s.gvr.String(), opts.Path, true, labels.Everything())
	if err != nil {
		return err
	}

	var sts appsv1.StatefulSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &sts)
	if err != nil {
		return errors.New("expecting StatefulSet resource")
	}
	if sts.Spec.Selector == nil || len(sts.Spec.Selector.MatchLabels) == 0 {
		return fmt.Errorf("No valid selector found on StatefulSet %s", opts.Path)
	}

	return podLogs(ctx, c, sts.Spec.Selector.MatchLabels, opts)
}
