// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/derailed/k9s/internal/slogs"
)

const (
	defaultBaseURL = "https://api.openai.com/v1"
	defaultModel   = "gpt-4"
	defaultTimeout = 30 * time.Second
)

// OpenAIProvider implements LLMProvider for OpenAI API.
type OpenAIProvider struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// NewOpenAIProvider creates a new OpenAI provider.
func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		baseURL: defaultBaseURL,
		client: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

// Name returns the provider name.
func (p *OpenAIProvider) Name() string {
	return "openai"
}

// IsConfigured returns true if the provider has an API key.
func (p *OpenAIProvider) IsConfigured() bool {
	return p.apiKey != ""
}

// Configure sets up the provider with the given configuration.
func (p *OpenAIProvider) Configure(config *ProviderConfig) error {
	if config.APIKey == "" {
		return fmt.Errorf("API key is required for OpenAI provider")
	}

	p.apiKey = config.APIKey
	if config.BaseURL != "" {
		p.baseURL = config.BaseURL
	}

	if config.Options != nil && config.Options.Timeout > 0 {
		p.client.Timeout = config.Options.Timeout
	}

	return nil
}

// openAIRequest represents the request format for OpenAI API.
type openAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// openAIResponse represents the response format from OpenAI API.
type openAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
	Usage UsageInfo `json:"usage"`
}

// SendMessage sends a message to OpenAI and returns the response.
func (p *OpenAIProvider) SendMessage(ctx context.Context, messages []Message, opts *Options) (*Response, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("OpenAI provider not configured")
	}

	req := openAIRequest{
		Model:    defaultModel,
		Messages: messages,
	}

	if opts != nil {
		if opts.Model != "" {
			req.Model = opts.Model
		}
		if opts.Temperature > 0 {
			req.Temperature = opts.Temperature
		}
		if opts.MaxTokens > 0 {
			req.MaxTokens = opts.MaxTokens
		}
		if opts.Timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
			defer cancel()
		}
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", p.baseURL+"/chat/completions", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	slog.Debug("Sending request to OpenAI", 
		slogs.Component, "chat",
		"operation", "send_message",
		"model", req.Model,
		"message_count", len(messages),
	)

	resp, err := p.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in OpenAI response")
	}

	return &Response{
		Content: openAIResp.Choices[0].Message.Content,
		Model:   openAIResp.Model,
		Usage:   openAIResp.Usage,
	}, nil
}

// openAIModelsResponse represents the models list response from OpenAI.
type openAIModelsResponse struct {
	Object string  `json:"object"`
	Data   []Model `json:"data"`
}

// GetModels returns available models from OpenAI.
func (p *OpenAIProvider) GetModels(ctx context.Context) ([]Model, error) {
	if !p.IsConfigured() {
		return nil, fmt.Errorf("OpenAI provider not configured")
	}

	req, err := http.NewRequestWithContext(ctx, "GET", p.baseURL+"/models", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get models: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	var modelsResp openAIModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to decode models response: %w", err)
	}

	return modelsResp.Data, nil
}
