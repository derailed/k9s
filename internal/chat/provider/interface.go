// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package provider

import (
	"context"
	"time"
)

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Model represents an available LLM model.
type Model struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// Options contains request options for LLM providers.
type Options struct {
	Model       string  `json:"model,omitempty"`
	Temperature float64 `json:"temperature,omitempty"`
	MaxTokens   int     `json:"max_tokens,omitempty"`
	Timeout     time.Duration
}

// Response represents an LLM response.
type Response struct {
	Content string    `json:"content"`
	Model   string    `json:"model"`
	Usage   UsageInfo `json:"usage"`
}

// UsageInfo contains token usage information.
type UsageInfo struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ProviderConfig contains provider-specific configuration.
type ProviderConfig struct {
	APIKey  string
	BaseURL string
	Model   string
	Options *Options
}

// LLMProvider defines the interface for LLM providers.
type LLMProvider interface {
	// SendMessage sends a message to the LLM and returns the response.
	SendMessage(ctx context.Context, messages []Message, opts *Options) (*Response, error)

	// GetModels returns available models from the provider.
	GetModels(ctx context.Context) ([]Model, error)

	// Configure configures the provider with the given config.
	Configure(config *ProviderConfig) error

	// Name returns the provider name.
	Name() string

	// IsConfigured returns true if the provider is properly configured.
	IsConfigured() bool
}
