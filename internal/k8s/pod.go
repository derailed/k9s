package k8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
)

const defaultKillGrace int64 = 5

type (
	// Loggable represents a K8s resource that has containers and can be logged.
	Loggable interface {
		Res
		Containers(ns, n string, includeInit bool) ([]string, error)
		Logs(ns, n, co string, lines int64, previous bool) *restclient.Request
	}

	// Pod represents a Kubernetes resource.
	Pod struct{}
)

// NewPod returns a new Pod.
func NewPod() Res {
	return &Pod{}
}

// Get a service.
func (*Pod) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().CoreV1().Pods(ns).Get(n, opts)
}

// List all services in a given namespace
func (*Pod) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().CoreV1().Pods(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a service
func (*Pod) Delete(ns, n string) error {
	var grace = defaultKillGrace
	opts := metav1.DeleteOptions{
		GracePeriodSeconds: &grace,
	}
	return conn.dialOrDie().CoreV1().Pods(ns).Delete(n, &opts)
}

// Containers returns all container names on pod
func (*Pod) Containers(ns, n string, includeInit bool) ([]string, error) {
	opts := metav1.GetOptions{}
	cc := []string{}
	po, err := conn.dialOrDie().CoreV1().Pods(ns).Get(n, opts)
	if err != nil {
		return cc, err
	}

	for _, c := range po.Spec.Containers {
		cc = append(cc, c.Name)
	}
	if includeInit {
		for _, c := range po.Spec.InitContainers {
			cc = append(cc, c.Name)
		}
	}
	return cc, nil
}

// Logs fetch container logs for a given pod and container.
func (*Pod) Logs(ns, n, co string, lines int64, prev bool) *restclient.Request {
	opts := &v1.PodLogOptions{
		Container: co,
		Follow:    true,
		TailLines: &lines,
		Previous:  prev,
	}

	return conn.dialOrDie().CoreV1().Pods(ns).GetLogs(n, opts)
}

// Events retrieved pod's events.
func (*Pod) Events(ns, n string) (*v1.EventList, error) {
	e := conn.dialOrDie().Core().Events(ns)
	sel := e.GetFieldSelector(&n, &ns, nil, nil)
	opts := metav1.ListOptions{FieldSelector: sel.String()}
	ee, err := e.List(opts)
	return ee, err
}
