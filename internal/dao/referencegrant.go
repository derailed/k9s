// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

var (
	_ Accessor = (*ReferenceGrant)(nil)
)

// ReferenceGrant represents a Kubernetes Gateway API ReferenceGrant resource.
type ReferenceGrant struct {
	Resource
}