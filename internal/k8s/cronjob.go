package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CronJob represents a Kubernetes CronJob
type CronJob struct{}

// NewCronJob returns a new CronJob.
func NewCronJob() Res {
	return &CronJob{}
}

// Get a CronJob.
func (c *CronJob) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().BatchV1beta1().CronJobs(ns).Get(n, opts)
}

// List all CronJobs in a given namespace
func (c *CronJob) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().BatchV1beta1().CronJobs(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a CronJob
func (c *CronJob) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().BatchV1beta1().CronJobs(ns).Delete(n, &opts)
}
