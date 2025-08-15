// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"time"

	"github.com/derailed/k9s/internal/chat/provider"
)

// MessageType represents the type of a chat message.
type MessageType int

const (
	MessageTypeUser MessageType = iota
	MessageTypeAssistant
	MessageTypeSystem
	MessageTypeError
)

// String returns the string representation of the message type.
func (mt MessageType) String() string {
	switch mt {
	case MessageTypeUser:
		return "user"
	case MessageTypeAssistant:
		return "assistant"
	case MessageTypeSystem:
		return "system"
	case MessageTypeError:
		return "error"
	default:
		return "unknown"
	}
}

// ChatMessage represents a single chat message.
type ChatMessage struct {
	ID        string
	Type      MessageType
	Content   string
	Timestamp time.Time
	Usage     *provider.UsageInfo
	Commands  []DetectedCommand
}

// DetectedCommand represents a command detected in a message.
type DetectedCommand struct {
	Command     string
	Description string
	RiskLevel   RiskLevel
	Approved    bool
}

// RiskLevel represents the risk level of a command.
type RiskLevel int

const (
	RiskLow RiskLevel = iota
	RiskMedium
	RiskHigh
)

// String returns the string representation of the risk level.
func (rl RiskLevel) String() string {
	switch rl {
	case RiskLow:
		return "low"
	case RiskMedium:
		return "medium"
	case RiskHigh:
		return "high"
	default:
		return "unknown"
	}
}

// Color returns the color for the risk level.
func (rl RiskLevel) Color() string {
	switch rl {
	case RiskLow:
		return "green"
	case RiskMedium:
		return "yellow"
	case RiskHigh:
		return "red"
	default:
		return "white"
	}
}

// NewChatMessage creates a new chat message.
func NewChatMessage(msgType MessageType, content string) *ChatMessage {
	return &ChatMessage{
		ID:        generateMessageID(),
		Type:      msgType,
		Content:   content,
		Timestamp: time.Now(),
	}
}

// ToProviderMessage converts a ChatMessage to a provider.Message.
func (m *ChatMessage) ToProviderMessage() provider.Message {
	return provider.Message{
		Role:    m.Type.String(),
		Content: m.Content,
	}
}

// generateMessageID generates a unique message ID.
func generateMessageID() string {
	return time.Now().Format("20060102150405.000000")
}
