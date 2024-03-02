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
