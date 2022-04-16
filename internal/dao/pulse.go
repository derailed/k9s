// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/runtime"
)

// Pulse tracks pulses.
type Pulse struct {
	NonResource
}

func (h *Pulse) Create(ctx context.Context, ns string, _ runtime.Object) (runtime.Object, error) {
	panic("Pulse create NYI")
}

// List lists out pulses.
func (h *Pulse) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	return nil, fmt.Errorf("NYI")
}
