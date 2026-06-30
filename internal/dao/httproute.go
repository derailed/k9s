// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

var (
	_ Accessor = (*HTTPRoute)(nil)
)

// HTTPRoute represents a Kubernetes Gateway API HTTPRoute resource.
type HTTPRoute struct {
	Resource
}