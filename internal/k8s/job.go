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
		Connection
	}

	// Loggable represents a K8s resource that has containers and can be logged.
	Loggable interface {
		Containers(ns, n string, includeInit bool) ([]string, error)
		Logs(ns, n, co string, lines int64, previous bool) *restclient.Request
	}
)

// NewJob returns a new Job.
func NewJob(c Connection) Cruder {
	return &Job{c}
}

// Get a Job.
func (j *Job) Get(ns, n string) (interface{}, error) {
	return j.DialOrDie().BatchV1().Jobs(ns).Get(n, metav1.GetOptions{})
}

// List all Jobs in a given namespace.
func (j *Job) List(ns string) (Collection, error) {
	rr, err := j.DialOrDie().BatchV1().Jobs(ns).List(metav1.ListOptions{})
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
func (j *Job) Delete(ns, n string) error {
	return j.DialOrDie().BatchV1().Jobs(ns).Delete(n, nil)
}

// Containers returns all container names on job.
func (j *Job) Containers(ns, n string, includeInit bool) ([]string, error) {
	pod, err := j.assocPod(ns, n)
	if err != nil {
		return nil, err
	}
	return NewPod(j).(Loggable).Containers(ns, pod, includeInit)
}

// Logs fetch container logs for a given job and container.
func (j *Job) Logs(ns, n, co string, lines int64, prev bool) *restclient.Request {
	pod, err := j.assocPod(ns, n)
	if err != nil {
		return nil
	}
	return NewPod(j).(Loggable).Logs(ns, pod, co, lines, prev)
}

// Events retrieved jobs events.
func (j *Job) Events(ns, n string) (*v1.EventList, error) {
	e := j.DialOrDie().Core().Events(ns)
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
