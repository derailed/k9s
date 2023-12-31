// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

// FeatureGates represents K9s opt-in features.
type FeatureGates struct {
	NodeShell bool `yaml:"nodeShell"`
}

// NewFeatureGates returns a new feature gate.
func NewFeatureGates() FeatureGates {
	return FeatureGates{}
}
