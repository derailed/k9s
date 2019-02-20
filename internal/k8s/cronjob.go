package k8s

import (
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

const maxJobNameSize = 42

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

// Run the job associated with this cronjob.
func (c *CronJob) Run(ns, n string) error {
	i, err := c.Get(ns, n)
	if err != nil {
		return err
	}

	cronJob := i.(*batchv1beta1.CronJob)

	var jobName = cronJob.Name
	if len(cronJob.Name) >= maxJobNameSize {
		jobName = cronJob.Name[0:41]
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName + "-manual-" + rand.String(3),
			Namespace: ns,
			Labels:    cronJob.Spec.JobTemplate.Labels,
		},
		Spec: cronJob.Spec.JobTemplate.Spec,
	}
	_, err = conn.dialOrDie().BatchV1().Jobs(ns).Create(job)
	return err
}
