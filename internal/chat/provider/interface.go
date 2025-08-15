// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package provider

import (
	"context"
	"time"

	"github.com/openai/openai-go/v2"
	"github.com/openai/openai-go/v2/shared"
)

// ToolExecutor defines the interface for kubectl command execution.
type ToolExecutor interface {
	// Execute runs a kubectl command and returns output.
	Execute(command string) (string, error)

	// NeedsApproval checks if a command requires user approval.
	NeedsApproval(command string) bool

	// RequestApproval requests user approval for a command.
	RequestApproval(command string, callback func(bool))
}

// K9sContext provides current k9s state information.
type K9sContext interface {
	// UpdateFromApp refreshes context from current k9s state.
	UpdateFromApp()

	// GetSystemPrompt returns the system prompt with current context.
	GetSystemPrompt() string

	// GetContextSummary returns a brief context summary.
	GetContextSummary() string
}

// ChatCompletionRequest represents a unified chat completion request.
type ChatCompletionRequest struct {
	Model       string                         `json:"model"`
	Messages    []openai.ChatCompletionMessage `json:"messages"`
	Tools       []openai.ChatCompletionTool    `json:"tools,omitempty"`
	ToolChoice  interface{}                    `json:"tool_choice,omitempty"`
	Temperature float64                        `json:"temperature,omitempty"`
	MaxTokens   int                            `json:"max_tokens,omitempty"`
}

// ChatCompletionResponse represents a unified chat completion response.
type ChatCompletionResponse struct {
	ID      string                        `json:"id"`
	Object  string                        `json:"object"`
	Created int64                         `json:"created"`
	Model   string                        `json:"model"`
	Choices []openai.ChatCompletionChoice `json:"choices"`
	Usage   openai.CompletionUsage        `json:"usage"`
}

// ProviderConfig contains provider-specific configuration.
type ProviderConfig struct {
	APIKey   string
	BaseURL  string
	Model    string
	Timeout  time.Duration
	Executor ToolExecutor
	Context  K9sContext
}

// LLMProvider defines the interface for LLM providers using openai-go types.
type LLMProvider interface {
	// CreateChatCompletion creates a chat completion using openai-go types.
	CreateChatCompletion(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error)

	// GetModels returns available models.
	GetModels(ctx context.Context) ([]shared.Model, error)

	// Configure configures the provider with the given config.
	Configure(config *ProviderConfig) error

	// Name returns the provider name.
	Name() string

	// IsConfigured returns true if the provider is properly configured.
	IsConfigured() bool
}
