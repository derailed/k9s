package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HPA represents am HorizontalPodAutoscaler.
type HPA struct{}

// NewHPA returns a new HPA.
func NewHPA() Res {
	return &HPA{}
}

// Get a service.
func (*HPA) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().AutoscalingV2beta2().HorizontalPodAutoscalers(ns).Get(n, opts)
}

// List all services in a given namespace
func (*HPA) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().AutoscalingV2beta2().HorizontalPodAutoscalers(ns).List(opts)
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
func (*HPA) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().Autoscaling().HorizontalPodAutoscalers(ns).Delete(n, &opts)
}
