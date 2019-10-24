package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HorizontalPodAutoscalerV1 represents am HorizontalPodAutoscaler.
type HorizontalPodAutoscalerV1 struct {
	*base
	Connection
}

// NewHorizontalPodAutoscalerV1 returns a new HorizontalPodAutoscaler.
func NewHorizontalPodAutoscalerV1(c Connection) *HorizontalPodAutoscalerV1 {
	return &HorizontalPodAutoscalerV1{&base{}, c}
}

// Get a HorizontalPodAutoscaler.
func (h *HorizontalPodAutoscalerV1) Get(ns, n string) (interface{}, error) {
	return h.DialOrDie().AutoscalingV1().HorizontalPodAutoscalers(ns).Get(n, metav1.GetOptions{})
}

// List all HorizontalPodAutoscalers in a given namespace.
func (h *HorizontalPodAutoscalerV1) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: h.labelSelector,
		FieldSelector: h.fieldSelector,
	}
	rr, err := h.DialOrDie().AutoscalingV1().HorizontalPodAutoscalers(ns).List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a HorizontalPodAutoscaler.
func (h *HorizontalPodAutoscalerV1) Delete(ns, n string, cascade, force bool) error {
	return h.DialOrDie().AutoscalingV1().HorizontalPodAutoscalers(ns).Delete(n, nil)
}
