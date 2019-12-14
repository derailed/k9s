package model

import (
	"fmt"

	"github.com/derailed/tview"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func extractFQN(o runtime.Object) string {
	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		log.Error().Err(fmt.Errorf("expecting unstructured but got %T", o))
		return "na"
	}
	m, ok := u.Object["metadata"].(map[string]interface{})
	if !ok {
		log.Error().Err(fmt.Errorf("expecting interface map for metadata but got %T", u.Object["metadata"]))
		return "na"
	}

	n, ok := m["name"].(string)
	if !ok {
		log.Error().Err(fmt.Errorf("expecting interface map for name but got %T", m["name"]))
		return "na"
	}

	ns, ok := m["namespace"].(string)
	if !ok {
		return FQN("", n)
	}

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
