// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*Pod)(nil)
)

// Namespace represents a namespace resource.
type Namespace struct {
	Generic
}

// List returns a collection of namespaces.
func (n *Namespace) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	oo, err := n.Generic.List(ctx, ns)
	if err != nil {
		return nil, err
	}

	return oo, nil
}
