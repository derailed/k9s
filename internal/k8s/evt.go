package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Event represents a Kubernetes Event.
type Event struct {
	*Resource
	Connection
}

// NewEvent returns a new Event.
func NewEvent(c Connection, gvr GVR) *Event {
	return &Event{&Resource{gvr: gvr}, c}
}

// Get a Event.
func (e *Event) Get(ns, n string) (interface{}, error) {
	return e.DialOrDie().CoreV1().Events(ns).Get(n, metav1.GetOptions{})
}

// List all Events in a given namespace.
func (e *Event) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: e.labelSelector,
		FieldSelector: e.fieldSelector,
	}
	rr, err := e.DialOrDie().CoreV1().Events(ns).List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete an Event.
func (e *Event) Delete(ns, n string, cascade, force bool) error {
	return e.DialOrDie().CoreV1().Events(ns).Delete(n, nil)
}
