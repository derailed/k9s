package dao

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

type Pulse struct {
	NonResource
}

func (h *Pulse) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	return nil, fmt.Errorf("NYI")
}
