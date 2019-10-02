package k8s

import (
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// HorizontalPodAutoscalerV1 represents am HorizontalPodAutoscaler.
type HorizontalPodAutoscalerV1 struct {
	*Resource
	Connection
}

// NewHorizontalPodAutoscalerV1 returns a new HorizontalPodAutoscaler.
func NewHorizontalPodAutoscalerV1(c Connection, gvr GVR) *HorizontalPodAutoscalerV1 {
	return &HorizontalPodAutoscalerV1{&Resource{gvr: gvr}, c}
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

	rr1, err := h.DialOrDie().AutoscalingV2beta2().HorizontalPodAutoscalers(ns).List(opts)
	if err != nil {
		return nil, err
	}
	for _, r := range rr1.Items {
		log.Debug().Msgf("MX %#v", len(r.Spec.Metrics))
		log.Debug().Msgf("HPA:%#v -- %s", r.TypeMeta.Kind, r.Name)
	}

	return cc, nil
}

// Delete a HorizontalPodAutoscaler.
func (h *HorizontalPodAutoscalerV1) Delete(ns, n string, cascade, force bool) error {
	return h.DialOrDie().AutoscalingV1().HorizontalPodAutoscalers(ns).Delete(n, nil)
}
