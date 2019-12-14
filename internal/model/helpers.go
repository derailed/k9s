package model

import (
	"github.com/derailed/tview"
	runewidth "github.com/mattn/go-runewidth"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func extractFQN(o runtime.Object) string {
	u := o.(*unstructured.Unstructured)
	m := u.Object["metadata"].(map[string]interface{})
	if _, ok := m["namespace"]; !ok {
		return FQN("", m["name"].(string))
	}
	ns, n := m["namespace"].(string), m["name"].(string)
	return FQN(ns, n)
}

// MetaFQN returns a fully qualified resource name.
func MetaFQN(m metav1.ObjectMeta) string {
	if m.Namespace == "" {
		return m.Name
	}

	return FQN(m.Namespace, m.Name)
}

// FQN returns a fully qualified resource name.
func FQN(ns, n string) string {
	if ns == "" {
		return n
	}
	return ns + "/" + n
}

// Truncate a string to the given l and suffix ellipsis if needed.
func Truncate(str string, width int) string {
	return runewidth.Truncate(str, width, string(tview.SemigraphicsHorizontalEllipsis))
}
