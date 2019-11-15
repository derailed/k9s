package k8s

import (
	"errors"

	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

const maxJobNameSize = 42

// CronJob represents a Kubernetes CronJob.
type CronJob struct {
	*base
	Connection
}

// NewCronJob returns a new CronJob.
func NewCronJob(c Connection) *CronJob {
	return &CronJob{&base{}, c}
}

// Get a CronJob.
func (c *CronJob) Get(ns, n string) (interface{}, error) {
	return c.DialOrDie().BatchV1beta1().CronJobs(ns).Get(n, metav1.GetOptions{})
}

// List all CronJobs in a given namespace.
func (c *CronJob) List(ns string, opts metav1.ListOptions) (Collection, error) {
	rr, err := c.DialOrDie().BatchV1beta1().CronJobs(ns).List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a CronJob.
func (c *CronJob) Delete(ns, n string, cascade, force bool) error {
	return c.DialOrDie().BatchV1beta1().CronJobs(ns).Delete(n, nil)
}

// Run the job associated with this cronjob.
func (c *CronJob) Run(ns, n string) error {
	cj, err := c.Get(ns, n)
	if err != nil {
		return err
	}

	cronJob, ok := cj.(*batchv1beta1.CronJob)
	if !ok {
		return errors.New("Expecting valid cronjob")
	}
	var jobName = cronJob.Name
	if len(cronJob.Name) >= maxJobNameSize {
		jobName = cronJob.Name[0:maxJobNameSize]
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName + "-manual-" + rand.String(3),
			Namespace: ns,
			Labels:    cronJob.Spec.JobTemplate.Labels,
		},
		Spec: cronJob.Spec.JobTemplate.Spec,
	}

	_, err = c.DialOrDie().BatchV1().Jobs(ns).Create(job)
	return err
}
