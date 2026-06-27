// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

var (
	_ Accessor = (*UDPRoute)(nil)
)

// UDPRoute represents a Kubernetes Gateway API UDPRoute resource.
type UDPRoute struct {
	Resource
}