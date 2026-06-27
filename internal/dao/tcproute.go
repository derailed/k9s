// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

var (
	_ Accessor = (*TCPRoute)(nil)
)

// TCPRoute represents a Kubernetes Gateway API TCPRoute resource.
type TCPRoute struct {
	Resource
}