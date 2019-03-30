package k8s

import (
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var supportedAutoScalingAPIVersions = []string{"v2beta2", "v2beta1", "v1"}

// HPAV2Beta2 represents am HorizontalPodAutoscaler.
type HPAV2Beta2 struct {
	Connection
}

// NewHPAV2Beta2 returns a new HPAV2Beta2.
func NewHPAV2Beta2(c Connection) Cruder {
	return &HPAV2Beta2{c}
}

// Get a HPAV2Beta2.
func (h *HPAV2Beta2) Get(ns, n string) (interface{}, error) {
	return h.DialOrDie().AutoscalingV2beta2().HorizontalPodAutoscalers(ns).Get(n, metav1.GetOptions{})
}

// List all HPAV2Beta2s in a given namespace.
func (h *HPAV2Beta2) List(ns string) (Collection, error) {
	log.Debug().Msg("!!!! YO V2B2")
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

// Delete a HPAV2Beta2.
func (h *HPAV2Beta2) Delete(ns, n string) error {
	return h.DialOrDie().AutoscalingV2beta2().HorizontalPodAutoscalers(ns).Delete(n, nil)
}
