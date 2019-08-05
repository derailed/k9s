package k8s

import (
	"fmt"
	"strings"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
)

type (
	// Job represents a Kubernetes Job.
	Job struct {
		*base
		Connection
	}

	// Loggable represents a K8s resource that has containers and can be logged.
	Loggable interface {
		Containers(ns, n string, includeInit bool) ([]string, error)
		Logs(ns, n string, opts *v1.PodLogOptions) *restclient.Request
	}
)

// NewJob returns a new Job.
func NewJob(c Connection) *Job {
	return &Job{&base{}, c}
}

// Get a Job.
func (j *Job) Get(ns, n string) (interface{}, error) {
	return j.DialOrDie().BatchV1().Jobs(ns).Get(n, metav1.GetOptions{})
}

// List all Jobs in a given namespace.
func (j *Job) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: j.labelSelector,
		FieldSelector: j.fieldSelector,
	}
	rr, err := j.DialOrDie().BatchV1().Jobs(ns).List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a Job.
func (j *Job) Delete(ns, n string, cascade, force bool) error {
	return j.DialOrDie().BatchV1().Jobs(ns).Delete(n, nil)
}

// Containers returns all container names on job.
func (j *Job) Containers(ns, n string, includeInit bool) ([]string, error) {
	pod, err := j.assocPod(ns, n)
	if err != nil {
		return nil, err
	}
	return NewPod(j).Containers(ns, pod, includeInit)
}

// Logs fetch container logs for a given job and container.
func (j *Job) Logs(ns, n string, opts *v1.PodLogOptions) *restclient.Request {
	pod, err := j.assocPod(ns, n)
	if err != nil {
		return nil
	}

	return NewPod(j).Logs(ns, pod, opts)
}

// Events retrieved jobs events.
func (j *Job) Events(ns, n string) (*v1.EventList, error) {
	e := j.DialOrDie().CoreV1().Events(ns)
	return e.List(metav1.ListOptions{
		FieldSelector: e.GetFieldSelector(&n, &ns, nil, nil).String(),
	})
}

func (j *Job) assocPod(ns, n string) (string, error) {
	ee, err := j.Events(ns, n)
	if err != nil {
		return "", err
	}

	for _, e := range ee.Items {
		if strings.Contains(e.Message, "Created pod: ") {
			return strings.TrimSpace(strings.Replace(e.Message, "Created pod: ", "", 1)), nil
		}
	}
	return "", fmt.Errorf("unable to find associated pod name for job: %s/%s", ns, n)
}
