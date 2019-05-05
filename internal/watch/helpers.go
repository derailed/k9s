package watch

import (
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
