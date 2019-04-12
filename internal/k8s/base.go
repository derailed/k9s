package k8s

type base struct {
	fieldSelector string
	labelSelector string
}

// SetFieldSelector refines query results via selector.
func (b *base) SetFieldSelector(s string) {
	b.fieldSelector = s
}

// SetLabelSelector refines query results via labels.
func (b *base) SetLabelSelector(s string) {
	b.labelSelector = s
}

func (b *base) HasSelectors() bool {
	return b.labelSelector != "" || b.fieldSelector != ""
}
