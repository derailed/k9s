// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import "os"

const (
	// DefaultAIModel is the default Claude model to use.
	DefaultAIModel = "claude-sonnet-4-20250514"
	// DefaultAIMaxTokens is the default max tokens for AI responses.
	DefaultAIMaxTokens = 4096
)

// AI tracks AI/Claude configuration options.
type AI struct {
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	APIKey    string `json:"apiKey,omitempty" yaml:"apiKey,omitempty"`
	APIKeyEnv string `json:"apiKeyEnv,omitempty" yaml:"apiKeyEnv,omitempty"`
	Model     string `json:"model,omitempty" yaml:"model,omitempty"`
	MaxTokens int    `json:"maxTokens,omitempty" yaml:"maxTokens,omitempty"`
}

// NewAI creates a new AI configuration with defaults.
func NewAI() AI {
	return AI{
		Enabled: false,
	}
}

// GetAPIKey returns the API key, checking config first, then env var.
func (a *AI) GetAPIKey() string {
	if a.APIKey != "" {
		return a.APIKey
	}
	if a.APIKeyEnv != "" {
		return os.Getenv(a.APIKeyEnv)
	}
	return os.Getenv("ANTHROPIC_API_KEY")
}

// GetModel returns the model to use, defaulting if not set.
func (a *AI) GetModel() string {
	if a.Model != "" {
		return a.Model
	}
	return DefaultAIModel
}

// GetMaxTokens returns the max tokens, defaulting if not set.
func (a *AI) GetMaxTokens() int {
	if a.MaxTokens > 0 {
		return a.MaxTokens
	}
	return DefaultAIMaxTokens
}
