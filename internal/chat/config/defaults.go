// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"time"

	"github.com/derailed/k9s/internal/chat/provider"
)

// Default configuration values
const (
	DefaultModel       = "gpt-4"
	DefaultTemperature = 0.7
	DefaultMaxTokens   = 2048
	DefaultTimeout     = 30 * time.Second
	DefaultMaxHistory  = 50
)

// GetDefaultOptions returns default provider options.
func GetDefaultOptions() *provider.Options {
	return &provider.Options{
		Model:       DefaultModel,
		Temperature: DefaultTemperature,
		MaxTokens:   DefaultMaxTokens,
		Timeout:     DefaultTimeout,
	}
}

// ValidateOptions validates and corrects provider options.
func ValidateOptions(opts *provider.Options) *provider.Options {
	if opts == nil {
		return GetDefaultOptions()
	}

	validated := &provider.Options{
		Model:       opts.Model,
		Temperature: opts.Temperature,
		MaxTokens:   opts.MaxTokens,
		Timeout:     opts.Timeout,
	}

	// Validate model
	if validated.Model == "" {
		validated.Model = DefaultModel
	}

	// Validate temperature (0.0 to 2.0)
	if validated.Temperature < 0 || validated.Temperature > 2 {
		validated.Temperature = DefaultTemperature
	}

	// Validate max tokens (must be positive)
	if validated.MaxTokens <= 0 {
		validated.MaxTokens = DefaultMaxTokens
	}

	// Validate timeout (must be positive)
	if validated.Timeout <= 0 {
		validated.Timeout = DefaultTimeout
	}

	return validated
}
