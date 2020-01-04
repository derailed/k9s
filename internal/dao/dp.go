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
	_ Accessor    = (*Deployment)(nil)
	_ Nuker       = (*Deployment)(nil)
	_ Loggable    = (*Deployment)(nil)
	_ Restartable = (*Deployment)(nil)
	_ Scalable    = (*Deployment)(nil)
)

// Deployment represents a deployment K8s resource.
type Deployment struct {
	Resource
}

// Scale a Deployment.
func (d *Deployment) Scale(path string, replicas int32) error {
	ns, n := client.Namespaced(path)
	auth, err := d.Client().CanI(ns, "apps/v1/deployments:scale", []string{"get", "update"})
	if !auth || err != nil {
		return err
	}

	scale, err := d.Client().DialOrDie().AppsV1().Deployments(ns).GetScale(n, metav1.GetOptions{})
	if err != nil {
		return err
	}
	scale.Spec.Replicas = replicas
	_, err = d.Client().DialOrDie().AppsV1().Deployments(ns).UpdateScale(n, scale)

	return err
}

// Restart a Deployment rollout.
func (d *Deployment) Restart(path string) error {
	o, err := d.Factory.Get(d.gvr.String(), path, true, labels.Everything())
	if err != nil {
		return err
	}
	var ds appsv1.Deployment
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ds)
	if err != nil {
		return err
	}

	ns, _ := client.Namespaced(path)
	auth, err := d.Client().CanI(ns, "apps/v1/deployments", []string{"patch"})
	if !auth || err != nil {
		return err
	}
	update, err := polymorphichelpers.ObjectRestarterFn(&ds)
	if err != nil {
		return err
	}
	_, err = d.Client().DialOrDie().AppsV1().Deployments(ds.Namespace).Patch(ds.Name, types.StrategicMergePatchType, update)

	return err
}

// TailLogs tail logs for all pods represented by this Deployment.
func (d *Deployment) TailLogs(ctx context.Context, c chan<- string, opts LogOptions) error {
	o, err := d.Factory.Get(d.gvr.String(), opts.Path, true, labels.Everything())
	if err != nil {
		return err
	}

	var dp appsv1.Deployment
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &dp)
	if err != nil {
		return errors.New("expecting Deployment resource")
	}
	if dp.Spec.Selector == nil || len(dp.Spec.Selector.MatchLabels) == 0 {
		return fmt.Errorf("No valid selector found on Deployment %s", opts.Path)
	}

	return podLogs(ctx, c, dp.Spec.Selector.MatchLabels, opts)
}
