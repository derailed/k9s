// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

var (
	_ Accessor = (*GRPCRoute)(nil)
)

// GRPCRoute represents a Kubernetes Gateway API GRPCRoute resource.
type GRPCRoute struct {
	Resource
}