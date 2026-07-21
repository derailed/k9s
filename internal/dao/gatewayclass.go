// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

var (
	_ Accessor = (*GatewayClass)(nil)
)

// GatewayClass represents a Kubernetes Gateway API GatewayClass resource.
type GatewayClass struct {
	Resource
}
