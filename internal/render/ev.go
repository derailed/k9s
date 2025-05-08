// Copyright Authors of K9s

package render

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/derailed/k9s/internal/slogs"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	api "k8s.io/kubernetes/pkg/apis/core"
)

// Event renders a event resource to screen.
type Event struct {
	Table
}

// Healthy checks component health.
func (*Event) Healthy(_ context.Context, o any) error {
	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		slog.Error("Expected Unstructured", slogs.Type, fmt.Sprintf("%T", o))
		return nil
	}
	var ev api.Event
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &ev)
	if err != nil {
		slog.Error("Failed to convert unstructured to Node", slogs.Error, err)
		return nil
	}
	if ev.Type != "Normal" {
		return fmt.Errorf("event is not normal: %s", ev.Type)
	}

	return nil
}
