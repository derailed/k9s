package k8s

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
)

// Job represents a Kubernetes Job
type Job struct{}

// NewJob returns a new Job.
func NewJob() Res {
	return &Job{}
}

// Get a Job.
func (*Job) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().BatchV1().Jobs(ns).Get(n, opts)
}

// List all Jobs in a given namespace
func (*Job) List(ns string) (Collection, error) {
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
func (*Job) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().BatchV1().Jobs(ns).Delete(n, &opts)
}

// Containers returns all container names on pod
func (j *Job) Containers(ns, n string, includeInit bool) ([]string, error) {
	pod, err := j.assocPod(ns, n)
	if err != nil {
		return nil, err
	}
	log.Debug("Containers found assoc pod", pod)
	return NewPod().(Loggable).Containers(ns, pod, includeInit)
}

// Logs fetch container logs for a given pod and container.
func (j *Job) Logs(ns, n, co string, lines int64, prev bool) *restclient.Request {
	pod, err := j.assocPod(ns, n)
	if err != nil {
		return nil
	}
	log.Println("Logs found assoc pod", pod)
	return NewPod().(Loggable).Logs(ns, pod, co, lines, prev)
}

// Events retrieved jobs events.
func (*Job) Events(ns, n string) (*v1.EventList, error) {
	e := conn.dialOrDie().Core().Events(ns)
	sel := e.GetFieldSelector(&n, &ns, nil, nil)
	opts := metav1.ListOptions{FieldSelector: sel.String()}
	ee, err := e.List(opts)
	return ee, err
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
