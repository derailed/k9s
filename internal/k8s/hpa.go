package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HPA represents am HorizontalPodAutoscaler.
type HPA struct {
	Connection
}

// NewHPA returns a new HPA.
func NewHPA(c Connection) Cruder {
	return &HPA{c}
}

// Get a HPA.
func (h *HPA) Get(ns, n string) (interface{}, error) {
	return h.DialOrDie().AutoscalingV2beta2().HorizontalPodAutoscalers(ns).Get(n, metav1.GetOptions{})
}

// List all HPAs in a given namespace.
func (h *HPA) List(ns string) (Collection, error) {
	rr, err := h.DialOrDie().AutoscalingV2beta2().HorizontalPodAutoscalers(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}
	return cc, nil
}

// Delete a HPA.
func (h *HPA) Delete(ns, n string) error {
	if h.SupportsResource("autoscaling/v2beta1") {
		return h.DialOrDie().AutoscalingV2beta1().HorizontalPodAutoscalers(ns).Delete(n, nil)
	}
	return h.DialOrDie().AutoscalingV2beta2().HorizontalPodAutoscalers(ns).Delete(n, nil)
}
