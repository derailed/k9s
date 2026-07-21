// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

var (
	_ Accessor = (*TLSRoute)(nil)
)

// TLSRoute represents a Kubernetes Gateway API TLSRoute resource.
type TLSRoute struct {
	Resource
}