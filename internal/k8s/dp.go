package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment represents a Kubernetes Deployment.
type Deployment struct {
	Connection
}

// NewDeployment returns a new Deployment.
func NewDeployment(c Connection) Cruder {
	return &Deployment{c}
}

// Get a deployment.
func (d *Deployment) Get(ns, n string) (interface{}, error) {
	return d.DialOrDie().Apps().Deployments(ns).Get(n, metav1.GetOptions{})
}

// List all Deployments in a given namespace.
func (d *Deployment) List(ns string) (Collection, error) {
	rr, err := d.DialOrDie().Apps().Deployments(ns).List(metav1.ListOptions{})
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
func (d *Deployment) Delete(ns, n string) error {
	return d.DialOrDie().Apps().Deployments(ns).Delete(n, nil)
}
