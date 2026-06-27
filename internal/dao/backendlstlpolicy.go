// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

var (
	_ Accessor = (*BackendTLSPolicy)(nil)
)

// BackendTLSPolicy represents a Kubernetes Gateway API BackendTLSPolicy resource.
type BackendTLSPolicy struct {
	Resource
}