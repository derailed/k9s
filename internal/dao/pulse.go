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

// List lists out pulses.
func (*Pulse) List(context.Context, string) ([]runtime.Object, error) {
	return nil, fmt.Errorf("NYI")
}
