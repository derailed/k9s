// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package chat

import (
	"time"

	"github.com/derailed/tview"
)

// MessageType represents the type of message.
type MessageType int

const (
	MessageTypeUser MessageType = iota
	MessageTypeAssistant
	MessageTypeSystem
	MessageTypeError
)

// Message represents a chat message.
type Message struct {
	ID        string                 `json:"id"`
	Content   string                 `json:"content"`
	Type      MessageType            `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// K9sContext represents the current k9s state.
type K9sContext struct {
	CurrentNamespace string
	CurrentView      string
	ClusterName      string
	SelectedItems    []string
	ActiveFilters    []string
}

// Provider interface for chat providers.
type Provider interface {
	GetResponse(message string, context K9sContext) (string, error)
}

// App interface for k9s app integration.
type App interface {
	QueueUpdateDraw(func())
	SetFocus(tview.Primitive)
	ToggleChat()
	SwitchFocus()
	GetCurrentNamespace() string
	GetCurrentView() string
	GetClusterName() string
}

// appWrapper implements App interface for k9s integration
type appWrapper struct {
	application *tview.Application
	factory     interface{} // watch.Factory, but we'll keep it generic for now
}

func (a *appWrapper) QueueUpdateDraw(f func()) {
	if a.application != nil {
		a.application.QueueUpdateDraw(f)
	}
}

func (a *appWrapper) SetFocus(p tview.Primitive) {
	if a.application != nil {
		a.application.SetFocus(p)
	}
}

func (a *appWrapper) ToggleChat() {
	// This will be called by the main app's toggle mechanism
}

func (a *appWrapper) SwitchFocus() {
	// This will be called by the main app's focus switching
}

func (a *appWrapper) GetCurrentNamespace() string {
	return "default" // Placeholder - will be properly implemented
}

func (a *appWrapper) GetCurrentView() string {
	return "main" // Placeholder - will be properly implemented
}

func (a *appWrapper) GetClusterName() string {
	return "k9s-cluster" // Placeholder - will be properly implemented
}
