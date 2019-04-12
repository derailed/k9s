package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HorizontalPodAutoscalerV2Beta1 represents am HorizontalPodAutoscaler.
type HorizontalPodAutoscalerV2Beta1 struct {
	*base
	Connection
}

// NewHorizontalPodAutoscalerV2Beta1 returns a new HorizontalPodAutoscaler.
func NewHorizontalPodAutoscalerV2Beta1(c Connection) *HorizontalPodAutoscalerV2Beta1 {
	return &HorizontalPodAutoscalerV2Beta1{&base{}, c}
}

// Get a HorizontalPodAutoscaler.
func (h *HorizontalPodAutoscalerV2Beta1) Get(ns, n string) (interface{}, error) {
	return h.DialOrDie().AutoscalingV2beta2().HorizontalPodAutoscalers(ns).Get(n, metav1.GetOptions{})
}

// List all HorizontalPodAutoscalers in a given namespace.
func (h *HorizontalPodAutoscalerV2Beta1) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: h.labelSelector,
		FieldSelector: h.fieldSelector,
	}
	rr, err := h.DialOrDie().AutoscalingV2beta2().HorizontalPodAutoscalers(ns).List(opts)
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
func (h *HorizontalPodAutoscalerV2Beta1) Delete(ns, n string) error {
	return h.DialOrDie().AutoscalingV2beta1().HorizontalPodAutoscalers(ns).Delete(n, nil)
}
