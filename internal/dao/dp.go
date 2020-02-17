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
	_ Controller  = (*Deployment)(nil)
)

// Deployment represents a deployment K8s resource.
type Deployment struct {
	Resource
}

// Scale a Deployment.
func (d *Deployment) Scale(path string, replicas int32) error {
	ns, n := client.Namespaced(path)
	auth, err := d.Client().CanI(ns, "apps/v1/deployments:scale", []string{client.GetVerb, client.UpdateVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to scale a deployment")
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
	dp, err := d.GetInstance(path)
	if err != nil {
		return err
	}

	ns, _ := client.Namespaced(path)
	auth, err := d.Client().CanI(ns, "apps/v1/deployments", []string{client.PatchVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to restart a deployment")
	}
	update, err := polymorphichelpers.ObjectRestarterFn(dp)
	if err != nil {
		return err
	}

	_, err = d.Client().DialOrDie().AppsV1().Deployments(dp.Namespace).Patch(dp.Name, types.StrategicMergePatchType, update)
	return err
}

// TailLogs tail logs for all pods represented by this Deployment.
func (d *Deployment) TailLogs(ctx context.Context, c chan<- []byte, opts LogOptions) error {
	dp, err := d.GetInstance(opts.Path)
	if err != nil {
		return err
	}
	if dp.Spec.Selector == nil || len(dp.Spec.Selector.MatchLabels) == 0 {
		return fmt.Errorf("No valid selector found on Deployment %s", opts.Path)
	}

	return podLogs(ctx, c, dp.Spec.Selector.MatchLabels, opts)
}

// Pod returns a pod victim by name.
func (d *Deployment) Pod(fqn string) (string, error) {
	dp, err := d.GetInstance(fqn)
	if err != nil {
		return "", err
	}

	return podFromSelector(d.Factory, dp.Namespace, dp.Spec.Selector.MatchLabels)
}

// GetInstance returns a deployment instance.
func (d *Deployment) GetInstance(fqn string) (*appsv1.Deployment, error) {
	o, err := d.Factory.Get(d.gvr.String(), fqn, false, labels.Everything())
	if err != nil {
		return nil, err
	}

	var dp appsv1.Deployment
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &dp)
	if err != nil {
		return nil, errors.New("expecting Deployment resource")
	}

	return &dp, nil
}
