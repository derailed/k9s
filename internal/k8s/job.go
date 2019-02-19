package k8s

import (
	batchv1 "k8s.io/api/batch/v1"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/rand"
)

// Job represents a Kubernetes Job
type Job struct{}

// NewJob returns a new Job.
func NewJob() Res {
	return &Job{}
}

func CreateJobFromCronJob(ns string, cj *batchv1beta1.CronJob) error {
	var jobName string
	if len(cj.Name) < 42 {
		jobName = cj.Name + "-manual-" + rand.String(3)
	} else {
		jobName = cj.Name[0:41] + "-manual-" + rand.String(3)
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: ns,
			Labels:    cj.Spec.JobTemplate.Labels,
		},
		Spec: cj.Spec.JobTemplate.Spec,
	}

	_, err := conn.dialOrDie().BatchV1().Jobs(ns).Create(job)
	return err
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
