package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var supportedAutoScalingAPIVersions = []string{"v2beta2", "v2beta1", "v1"}

// HorizontalPodAutoscalerV2Beta2 represents am HorizontalPodAutoscaler.
type HorizontalPodAutoscalerV2Beta2 struct {
	*base
	Connection
}

// NewHorizontalPodAutoscalerV2Beta2 returns a new HorizontalPodAutoscalerV2Beta2.
func NewHorizontalPodAutoscalerV2Beta2(c Connection) *HorizontalPodAutoscalerV2Beta2 {
	return &HorizontalPodAutoscalerV2Beta2{&base{}, c}
}

// Get a HorizontalPodAutoscalerV2Beta2.
func (h *HorizontalPodAutoscalerV2Beta2) Get(ns, n string) (interface{}, error) {
	return h.DialOrDie().AutoscalingV2beta2().HorizontalPodAutoscalers(ns).Get(n, metav1.GetOptions{})
}

// List all HorizontalPodAutoscalerV2Beta2s in a given namespace.
func (h *HorizontalPodAutoscalerV2Beta2) List(ns string) (Collection, error) {
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

// Delete a HorizontalPodAutoscalerV2Beta2.
func (h *HorizontalPodAutoscalerV2Beta2) Delete(ns, n string) error {
	return h.DialOrDie().AutoscalingV2beta2().HorizontalPodAutoscalers(ns).Delete(n, nil)
}
