// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"sync"
	"time"

	"github.com/derailed/k9s/internal/chat/provider"
)

// ChatState manages the persistent state of the chat feature.
type ChatState struct {
	mx          sync.RWMutex
	messages    []*ChatMessage
	provider    provider.LLMProvider
	k9sContext  *K9sContext
	config      *ChatConfig
	visible     bool
	initialized bool
}

// ChatConfig contains configuration for the chat feature.
type ChatConfig struct {
	Provider    string
	Model       string
	Temperature float64
	MaxTokens   int
	Timeout     time.Duration
	MaxHistory  int
}

// DefaultChatConfig returns default configuration.
func DefaultChatConfig() *ChatConfig {
	return &ChatConfig{
		Provider:    "openai",
		Model:       "gpt-4",
		Temperature: 0.7,
		MaxTokens:   2048,
		Timeout:     30 * time.Second,
		MaxHistory:  50,
	}
}

// NewChatState creates a new chat state.
func NewChatState() *ChatState {
	return &ChatState{
		messages:   make([]*ChatMessage, 0),
		k9sContext: NewK9sContext(),
		config:     DefaultChatConfig(),
		visible:    false,
	}
}

// IsVisible returns true if chat is currently visible.
func (s *ChatState) IsVisible() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.visible
}

// SetVisible sets the visibility of the chat.
func (s *ChatState) SetVisible(visible bool) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.visible = visible
}

// IsInitialized returns true if chat is initialized.
func (s *ChatState) IsInitialized() bool {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.initialized
}

// SetInitialized sets the initialization state.
func (s *ChatState) SetInitialized(initialized bool) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.initialized = initialized
}

// GetProvider returns the current LLM provider.
func (s *ChatState) GetProvider() provider.LLMProvider {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.provider
}

// SetProvider sets the LLM provider.
func (s *ChatState) SetProvider(p provider.LLMProvider) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.provider = p
}

// GetK9sContext returns the current k9s context.
func (s *ChatState) GetK9sContext() *K9sContext {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.k9sContext
}

// UpdateK9sContext updates the k9s context.
func (s *ChatState) UpdateK9sContext(clusterName, contextName, namespace, resource, view, selectedItem string) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.k9sContext.Update(clusterName, contextName, namespace, resource, view, selectedItem)
}

// GetConfig returns the current configuration.
func (s *ChatState) GetConfig() *ChatConfig {
	s.mx.RLock()
	defer s.mx.RUnlock()
	return s.config
}

// SetConfig sets the configuration.
func (s *ChatState) SetConfig(config *ChatConfig) {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.config = config
}

// AddMessage adds a message to the history.
func (s *ChatState) AddMessage(msg *ChatMessage) {
	s.mx.Lock()
	defer s.mx.Unlock()

	s.messages = append(s.messages, msg)

	// Limit message history
	if len(s.messages) > s.config.MaxHistory {
		s.messages = s.messages[len(s.messages)-s.config.MaxHistory:]
	}
}

// GetMessages returns a copy of the message history.
func (s *ChatState) GetMessages() []*ChatMessage {
	s.mx.RLock()
	defer s.mx.RUnlock()

	messages := make([]*ChatMessage, len(s.messages))
	copy(messages, s.messages)
	return messages
}

// ClearMessages clears the message history.
func (s *ChatState) ClearMessages() {
	s.mx.Lock()
	defer s.mx.Unlock()
	s.messages = make([]*ChatMessage, 0)
}
