// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"os"
	"strconv"
	"time"

	"github.com/derailed/k9s/internal/chat/provider"
)

const (
	// Environment variables
	EnvOpenAIAPIKey    = "OPENAI_API_KEY"
	EnvOpenAIBaseURL   = "OPENAI_BASE_URL"
	EnvChatModel       = "K9S_CHAT_MODEL"
	EnvChatTemperature = "K9S_CHAT_TEMPERATURE"
	EnvChatMaxTokens   = "K9S_CHAT_MAX_TOKENS"
	EnvChatTimeout     = "K9S_CHAT_TIMEOUT"
)

// LoadFromEnv loads configuration from environment variables.
func LoadFromEnv() *provider.ProviderConfig {
	config := &provider.ProviderConfig{
		APIKey:  os.Getenv(EnvOpenAIAPIKey),
		BaseURL: os.Getenv(EnvOpenAIBaseURL),
		Model:   getEnvOrDefault(EnvChatModel, "gpt-4"),
		Options: &provider.Options{},
	}

	// Parse temperature
	if tempStr := os.Getenv(EnvChatTemperature); tempStr != "" {
		if temp, err := strconv.ParseFloat(tempStr, 64); err == nil && temp >= 0 && temp <= 2 {
			config.Options.Temperature = temp
		}
	}
	if config.Options.Temperature == 0 {
		config.Options.Temperature = 0.7
	}

	// Parse max tokens
	if tokensStr := os.Getenv(EnvChatMaxTokens); tokensStr != "" {
		if tokens, err := strconv.Atoi(tokensStr); err == nil && tokens > 0 {
			config.Options.MaxTokens = tokens
		}
	}
	if config.Options.MaxTokens == 0 {
		config.Options.MaxTokens = 2048
	}

	// Parse timeout
	if timeoutStr := os.Getenv(EnvChatTimeout); timeoutStr != "" {
		if timeout, err := time.ParseDuration(timeoutStr); err == nil {
			config.Options.Timeout = timeout
		}
	}
	if config.Options.Timeout == 0 {
		config.Options.Timeout = 30 * time.Second
	}

	return config
}

// IsConfigured checks if the minimum required configuration is present.
func IsConfigured() bool {
	return os.Getenv(EnvOpenAIAPIKey) != ""
}

// getEnvOrDefault returns the environment variable value or a default.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
