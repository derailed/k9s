// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

const (
	defaultRefreshRate  = 2
	defaultMaxConnRetry = 5
)

// UI tracks ui specific configs.
type UI struct {
	// EnableMouse toggles mouse support.
	EnableMouse bool `yaml:"enableMouse"`

	// Headless toggles top header display.
	Headless bool `yaml:"headless"`

	// LogoLess toggles k9s logo.
	Logoless bool `yaml:"logoless"`

	// Crumbsless toggles nav crumb display.
	Crumbsless bool `yaml:"crumbsless"`

	// NoIcons toggles icons display.
	NoIcons bool `yaml:"noIcons"`

	// Skin reference the general k9s skin name.
	// Can be overridden per context.
	Skin string `yaml:"skin,omitempty"`
}
