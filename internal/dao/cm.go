// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

var (
	_ Accessor = (*ConfigMap)(nil)
)

// ConfigMap represents a configmap resource.
type ConfigMap struct {
	Resource
}
