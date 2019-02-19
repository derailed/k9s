package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Deployment represents a Kubernetes Deployment
type Deployment struct{}

// NewDeployment returns a new Deployment.
func NewDeployment() Res {
	return &Deployment{}
}

// Get a service.
func (*Deployment) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().Apps().Deployments(ns).Get(n, opts)
}

// List all Deployments in a given namespace
func (*Deployment) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().Apps().Deployments(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a Deployment
func (*Deployment) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().Apps().Deployments(ns).Delete(n, &opts)
}
