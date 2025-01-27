// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

// Proxy tracks a context's proxy configuration.
type Proxy struct {
	Address string `yaml:"address"`
}
