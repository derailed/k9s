package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Job represents a Kubernetes Job
type Job struct{}

// NewJob returns a new Job.
func NewJob() Res {
	return &Job{}
}

// Get a Job.
func (c *Job) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().BatchV1().Jobs(ns).Get(n, opts)
}

// List all Jobs in a given namespace
func (c *Job) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().BatchV1().Jobs(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a Job
func (c *Job) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().BatchV1().Jobs(ns).Delete(n, &opts)
}
