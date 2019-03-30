package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HPAV2Beta1 represents am HorizontalPodAutoscaler.
type HPAV2Beta1 struct {
	Connection
}

// NewHPAV2Beta1 returns a new HPA.
func NewHPAV2Beta1(c Connection) Cruder {
	return &HPAV2Beta1{c}
}

// Get a HPA.
func (h *HPAV2Beta1) Get(ns, n string) (interface{}, error) {
	return h.DialOrDie().AutoscalingV2beta2().HorizontalPodAutoscalers(ns).Get(n, metav1.GetOptions{})
}

// List all HPAs in a given namespace.
func (h *HPAV2Beta1) List(ns string) (Collection, error) {
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
func (h *HPAV2Beta1) Delete(ns, n string) error {
	return h.DialOrDie().AutoscalingV2beta1().HorizontalPodAutoscalers(ns).Delete(n, nil)
}
