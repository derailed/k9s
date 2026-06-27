// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

var (
	_ Accessor = (*Gateway)(nil)
)

// Gateway represents a Kubernetes Gateway API Gateway resource.
type Gateway struct {
	Resource
}
