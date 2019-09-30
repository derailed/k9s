package k8s

type Resource struct {
	fieldSelector string
	labelSelector string
	gvr           GVR
}

func (r *Resource) GVR() GVR {
	return r.gvr
}

// SetFieldSelector refines query results via selector.
func (r *Resource) SetFieldSelector(s string) {
	r.fieldSelector = s
}

// SetLabelSelector refines query results via labels.
func (r *Resource) SetLabelSelector(s string) {
	r.labelSelector = s
}

// GetFieldSelector returns field selector.
func (r *Resource) GetFieldSelector() string {
	return r.fieldSelector
}

// GetLabelSelector returns label selector.
func (r *Resource) GetLabelSelector() string {
	return r.labelSelector
}

func (r *Resource) HasSelectors() bool {
	return r.labelSelector != "" || r.fieldSelector != ""
}

func (r *Resource) Kill(ns, n string) error {
	return nil
}
