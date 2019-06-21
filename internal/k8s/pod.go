package k8s

import (
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	restclient "k8s.io/client-go/rest"
)

const defaultKillGrace int64 = 5

// Pod represents a Kubernetes Pod.
type Pod struct {
	*base
	Connection
}

// NewPod returns a new Pod.
func NewPod(c Connection) *Pod {
	return &Pod{base: &base{}, Connection: c}
}

// Get a pod.
func (p *Pod) Get(ns, name string) (interface{}, error) {
	return p.DialOrDie().CoreV1().Pods(ns).Get(name, metav1.GetOptions{})
}

// List all pods in a given namespace.
func (p *Pod) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: p.labelSelector,
		FieldSelector: p.fieldSelector,
	}

	rr, err := p.DialOrDie().CoreV1().Pods(ns).List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a pod.
func (p *Pod) Delete(ns, n string, cascade, force bool) error {
	log.Debug().Msgf("Killing Pod %s %t:%t", n, cascade, force)
	grace := defaultKillGrace
	if force {
		grace = 0
	}
	return p.DialOrDie().CoreV1().Pods(ns).Delete(n, &metav1.DeleteOptions{
		GracePeriodSeconds: &grace,
	})
}

// Containers returns all container names on pod
func (p *Pod) Containers(ns, n string, includeInit bool) ([]string, error) {
	po, err := p.DialOrDie().CoreV1().Pods(ns).Get(n, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	cc := []string{}
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
func (p *Pod) Logs(ns, n string, opts *v1.PodLogOptions) *restclient.Request {
	log.Debug().Msgf("Log Options %#v", *opts.TailLines)
	return p.DialOrDie().CoreV1().Pods(ns).GetLogs(n, opts)
}
