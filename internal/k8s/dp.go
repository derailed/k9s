package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment represents a Kubernetes Deployment.
type Deployment struct {
	*base
	Connection
}

// NewDeployment returns a new Deployment.
func NewDeployment(c Connection) *Deployment {
	return &Deployment{&base{}, c}
}

// Get a deployment.
func (d *Deployment) Get(ns, n string) (interface{}, error) {
	return d.DialOrDie().Apps().Deployments(ns).Get(n, metav1.GetOptions{})
}

// List all Deployments in a given namespace.
func (d *Deployment) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: d.labelSelector,
		FieldSelector: d.fieldSelector,
	}

	rr, err := d.DialOrDie().Apps().Deployments(ns).List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a Deployment.
func (d *Deployment) Delete(ns, n string, cascade, force bool) error {
	return d.DialOrDie().Apps().Deployments(ns).Delete(n, nil)
}

// Scale a Deployment.
func (d *Deployment) Scale(ns, n string, replicas int32) error {
	scale, err := d.DialOrDie().Apps().Deployments(ns).GetScale(n, metav1.GetOptions{})
	if err != nil {
		return err
	}

	scale.Spec.Replicas = replicas
	_, err = d.DialOrDie().Apps().Deployments(ns).UpdateScale(n, scale)
	return err
}
