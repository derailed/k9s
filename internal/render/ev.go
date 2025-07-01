// Copyright Authors of K9s

package render

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/derailed/k9s/internal/slogs"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Event renders a event resource to screen.
type Event struct {
	Table
}

// Healthy checks component health.
func (*Event) Healthy(_ context.Context, o any) error {
	r, ok := o.(metav1.TableRow)
	if !ok {
		slog.Error("Expected TableRow", slogs.Type, fmt.Sprintf("%T", o))
		return nil
	}
	idx := 2
	if idx < len(r.Cells) && r.Cells[idx] != "Normal" {
		return fmt.Errorf("event is not normal: %s", r.Cells[idx])
	}

	return nil
}
