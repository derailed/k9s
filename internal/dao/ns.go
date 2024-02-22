// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

var (
	_ Accessor = (*Namespace)(nil)
)

// Namespace represents a namespace resource.
type Namespace struct {
	Resource
}

// // List returns a collection of namespaces.
// func (n *Namespace) List(ctx context.Context, ns string) ([]runtime.Object, error) {
// 	oo, err := n.Generic.List(ctx, ns)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return oo, nil
// }
