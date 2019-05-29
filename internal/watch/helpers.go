package watch

import (
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func resourceDiff(l1, l2 v1.ResourceList) bool {
	if l1.Cpu().Cmp(*l2.Cpu()) != 0 {
		return true
	}
	if l1.Memory().Cmp(*l2.Memory()) != 0 {
		return true
	}

	return false
}

// MetaFQN computes unique resource id based on metadata.
func MetaFQN(m metav1.ObjectMeta) string {
	if m.Namespace == "" {
		return m.Name
	}

	return m.Namespace + "/" + m.Name
}

// ToSelector converts a string selector into a map.
func toSelector(s string) map[string]string {
	var m map[string]string
	ls, err := metav1.ParseToLabelSelector(s)
	if err != nil {
		log.Error().Err(err).Msg("StringToSel")
		return m
	}
	mSel, err := metav1.LabelSelectorAsMap(ls)
	if err != nil {
		log.Error().Err(err).Msg("SelToMap")
		return m
	}

	return mSel
}

// MatchesNode checks if pod selector matches node name.
func matchesNode(name string, selector map[string]string) bool {
	if len(selector) == 0 {
		return true
	}

	return selector["spec.nodeName"] == name
}

// MatchesLabels check if pod labels matches a given selector.
func matchesLabels(labels, selector map[string]string) bool {
	if len(selector) == 0 {
		return true
	}
	for k, v := range selector {
		la, ok := labels[k]
		if !ok || la != v {
			return false
		}
	}

	return true
}
