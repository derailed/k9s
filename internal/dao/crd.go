// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

var (
	_ Accessor = (*CustomResourceDefinition)(nil)
	_ Nuker    = (*CustomResourceDefinition)(nil)
)

// CustomResourceDefinition represents a CRD resource model.
type CustomResourceDefinition struct {
	Resource
}
