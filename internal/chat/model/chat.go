// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
	"github.com/charmbracelet/lipgloss"
	"github.com/derailed/k9s/internal/chat/provider"
)

// ChatModel represents the main chat Bubble Tea model.
type ChatModel struct {
	// State
	messages    []*ChatMessage
	input       textinput.Model
	viewport    viewport.Model
	k9sContext  *K9sContext
	
	// Configuration
	width       int
	height      int
	focused     bool
	
	// Dependencies
	provider    provider.LLMProvider
	
	// Status
	loading     bool
	err         error
	
	// Styling
	styles      ChatStyles
	glamour     *glamour.TermRenderer
}

// ChatStyles contains styling definitions for the chat UI.
type ChatStyles struct {
	Border         lipgloss.Style
	Header         lipgloss.Style
	UserMessage    lipgloss.Style
	BotMessage     lipgloss.Style
	ErrorMessage   lipgloss.Style
	SystemMessage  lipgloss.Style
	InputPrompt    lipgloss.Style
	LoadingMessage lipgloss.Style
	ContextInfo    lipgloss.Style
	CodeBlock      lipgloss.Style
}

// NewChatStyles creates default chat styles.
func NewChatStyles() ChatStyles {
	border := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#3C4048"))

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7AA2F7")).
		Background(lipgloss.Color("#1A1B26")).
		Padding(0, 1).
		MarginBottom(1)

	userMsg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#BB9AF7")).
		Bold(true).
		MarginBottom(1)

	botMsg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#9ECE6A")).
		MarginBottom(1)

	errorMsg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#F7768E")).
		Bold(true).
		MarginBottom(1)

	systemMsg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#E0AF68")).
		Italic(true).
		MarginBottom(1)

	inputPrompt := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#7AA2F7")).
		Bold(true)

	loadingMsg := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF9E64")).
		Bold(true)

	contextInfo := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#565F89")).
		Italic(true).
		Align(lipgloss.Center)

	codeBlock := lipgloss.NewStyle().
		Background(lipgloss.Color("#2D3748")).
		Padding(1).
		MarginTop(1).
		MarginBottom(1)

	return ChatStyles{
		Border:         border,
		Header:         header,
		UserMessage:    userMsg,
		BotMessage:     botMsg,
		ErrorMessage:   errorMsg,
		SystemMessage:  systemMsg,
		InputPrompt:    inputPrompt,
		LoadingMessage: loadingMsg,
		ContextInfo:    contextInfo,
		CodeBlock:      codeBlock,
	}
}

// NewChatModel creates a new chat model.
func NewChatModel(p provider.LLMProvider, ctx *K9sContext) *ChatModel {
	input := textinput.New()
	input.Placeholder = "Ask me about Kubernetes..."
	input.Focus()
	input.CharLimit = 500
	input.Width = 50
	input.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#7AA2F7"))
	input.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#C0CAF5"))

	viewport := viewport.New(80, 20)
	
	styles := NewChatStyles()
	
	// Initialize glamour for markdown rendering
	glamourRenderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(78),
	)

	welcomeMsg := NewChatMessage(MessageTypeSystem, "🤖 **K9s Chat Assistant** ready! I can help you with Kubernetes operations.\n\nType your questions and I'll provide assistance based on your current k9s context.")

	m := &ChatModel{
		messages:   []*ChatMessage{welcomeMsg},
		input:      input,
		viewport:   viewport,
		k9sContext: ctx,
		provider:   p,
		focused:    true,
		styles:     styles,
		glamour:    glamourRenderer,
	}
	
	m.updateViewport()
	return m
}

// MessageType definitions for tea messages
type (
	// MsgLLMResponse represents an LLM response message.
	MsgLLMResponse struct {
		Response *provider.Response
		Error    error
	}

	// MsgSendMessage represents a message to send to LLM.
	MsgSendMessage struct {
		Content string
	}

	// MsgResize represents a resize event.
	MsgResize struct {
		Width  int
		Height int
	}
)

// Init initializes the chat model.
func (m ChatModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles Bubble Tea messages.
func (m ChatModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if !m.focused {
			return m, nil
		}

		switch msg.String() {
		case "enter":
			if !m.loading && strings.TrimSpace(m.input.Value()) != "" {
				content := strings.TrimSpace(m.input.Value())
				m.input.SetValue("")
				
				// Add user message
				userMsg := NewChatMessage(MessageTypeUser, content)
				m.messages = append(m.messages, userMsg)
				
				m.loading = true
				m.updateViewport()
				
				return m, m.sendToLLM(content)
			}
		case "ctrl+c":
			return m, tea.Quit
		}

	case MsgLLMResponse:
		m.loading = false
		if msg.Error != nil {
			m.err = msg.Error
			errorMsg := NewChatMessage(MessageTypeError, "Error: "+msg.Error.Error())
			m.messages = append(m.messages, errorMsg)
		} else {
			assistantMsg := NewChatMessage(MessageTypeAssistant, msg.Response.Content)
			assistantMsg.Usage = &msg.Response.Usage
			m.messages = append(m.messages, assistantMsg)
		}
		m.updateViewport()

	case MsgResize:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = msg.Width - 4
		m.viewport.Width = msg.Width - 2
		m.viewport.Height = msg.Height - 5 // Leave space for input
		m.updateViewport()

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.Width = msg.Width - 4
		m.viewport.Width = msg.Width - 2
		m.viewport.Height = msg.Height - 5
		m.updateViewport()
	}

	// Update input
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)

	// Update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

// View renders the chat model.
func (m ChatModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Initializing chat..."
	}

	var b strings.Builder

	// Header with context info
	header := m.renderHeader()
	b.WriteString(header)
	b.WriteString("\n")

	// Messages viewport
	b.WriteString(m.viewport.View())
	b.WriteString("\n")

	// Input area
	inputArea := m.renderInput()
	b.WriteString(inputArea)

	return b.String()
}

// renderHeader renders the chat header with context information.
func (m ChatModel) renderHeader() string {
	contextSummary := m.k9sContext.GetContextSummary()
	if contextSummary == "No context" {
		contextSummary = "🤖 K9s Chat Assistant"
	} else {
		contextSummary = fmt.Sprintf("🤖 K9s Chat | %s", contextSummary)
	}

	header := m.styles.Header.Width(m.width).Render(contextSummary)
	
	// Add context details if available
	var contextDetails []string
	if m.k9sContext.ClusterName != "" && m.k9sContext.ClusterName != "unknown" {
		contextDetails = append(contextDetails, fmt.Sprintf("📡 %s", m.k9sContext.ClusterName))
	}
	if m.k9sContext.Namespace != "" && m.k9sContext.Namespace != "default" {
		contextDetails = append(contextDetails, fmt.Sprintf("📦 %s", m.k9sContext.Namespace))
	}
	if m.k9sContext.CurrentView != "" && m.k9sContext.CurrentView != "unknown" {
		contextDetails = append(contextDetails, fmt.Sprintf("👁️  %s", m.k9sContext.CurrentView))
	}
	
	if len(contextDetails) > 0 {
		details := strings.Join(contextDetails, " • ")
		contextInfo := m.styles.ContextInfo.Width(m.width).Render(details)
		return header + "\n" + contextInfo
	}
	
	return header
}

// renderInput renders the input area.
func (m ChatModel) renderInput() string {
	var b strings.Builder

	if m.loading {
		spinner := "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"
		idx := int(time.Now().UnixMilli()/100) % len(spinner)
		loadingText := fmt.Sprintf("%c Thinking...", rune(spinner[idx]))
		b.WriteString(m.styles.LoadingMessage.Render(loadingText))
		b.WriteString("\n")
	} else if m.err != nil {
		errorText := fmt.Sprintf("❌ %s", m.err.Error())
		b.WriteString(m.styles.ErrorMessage.Render(errorText))
		b.WriteString("\n")
	}

	prompt := m.styles.InputPrompt.Render("❯ ")
	b.WriteString(prompt + m.input.View())

	// Add helpful hints at the bottom
	if !m.loading && m.err == nil && m.input.Value() == "" {
		hints := m.styles.ContextInfo.
			Width(m.width).
			Render("💡 Try: \"What pods are running?\" or \"Show me the current namespace\"")
		b.WriteString("\n\n" + hints)
	}

	return b.String()
}

// updateViewport updates the viewport with current messages.
func (m *ChatModel) updateViewport() {
	var content strings.Builder

	for _, msg := range m.messages {
		content.WriteString(m.renderMessage(msg))
		content.WriteString("\n\n")
	}

	m.viewport.SetContent(content.String())
	m.viewport.GotoBottom()
}

// renderMessage renders a single message with proper styling and markdown support.
func (m ChatModel) renderMessage(msg *ChatMessage) string {
	var prefix string
	var style lipgloss.Style
	content := msg.Content

	switch msg.Type {
	case MessageTypeUser:
		prefix = "👤 **You:** "
		style = m.styles.UserMessage
		content = prefix + content
	case MessageTypeAssistant:
		prefix = "🤖 **Assistant:** "
		style = m.styles.BotMessage
		content = prefix + content
		
		// Try to render as markdown for assistant messages
		if m.glamour != nil {
			if rendered, err := m.glamour.Render(content); err == nil {
				return style.Render(strings.TrimSpace(rendered))
			}
		}
	case MessageTypeError:
		prefix = "❌ **Error:** "
		style = m.styles.ErrorMessage
		content = prefix + content
	case MessageTypeSystem:
		style = m.styles.SystemMessage
		
		// Try to render as markdown for system messages
		if m.glamour != nil {
			if rendered, err := m.glamour.Render(content); err == nil {
				return style.Render(strings.TrimSpace(rendered))
			}
		}
	}

	// Wrap text if needed
	if len(content) > m.width-10 {
		content = wordWrap(content, m.width-10)
	}

	// Add timestamp for user and assistant messages
	if msg.Type == MessageTypeUser || msg.Type == MessageTypeAssistant {
		timestamp := msg.Timestamp.Format("15:04")
		timestampStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("#565F89")).
			Faint(true)
		content += "\n" + timestampStyle.Render(fmt.Sprintf("📅 %s", timestamp))
	}

	return style.Render(content)
}

// sendToLLM sends a message to the LLM provider.
func (m ChatModel) sendToLLM(content string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// Prepare messages for LLM
		var messages []provider.Message

		// Add system prompt with k9s context
		systemPrompt := m.k9sContext.GetSystemPrompt()
		messages = append(messages, provider.Message{
			Role:    "system",
			Content: systemPrompt,
		})

		// Add recent message history (last 10 messages)
		start := 0
		if len(m.messages) > 10 {
			start = len(m.messages) - 10
		}

		for i := start; i < len(m.messages); i++ {
			if m.messages[i].Type != MessageTypeError {
				messages = append(messages, m.messages[i].ToProviderMessage())
			}
		}

		// Add current user message
		messages = append(messages, provider.Message{
			Role:    "user",
			Content: content,
		})

		response, err := m.provider.SendMessage(ctx, messages, &provider.Options{
			Temperature: 0.7,
			MaxTokens:   2048,
			Timeout:     30 * time.Second,
		})

		return MsgLLMResponse{
			Response: response,
			Error:    err,
		}
	}
}

// SetFocus sets the focus state of the chat model.
func (m *ChatModel) SetFocus(focused bool) {
	m.focused = focused
	if focused {
		m.input.Focus()
	} else {
		m.input.Blur()
	}
}

// SetSize sets the size of the chat model.
func (m *ChatModel) SetSize(width, height int) {
	m.width = width
	m.height = height
	m.input.Width = width - 4
	m.viewport.Width = width - 2
	m.viewport.Height = height - 5
	m.updateViewport()
}

// AddMessage adds a new message to the chat.
func (m *ChatModel) AddMessage(msgType MessageType, content string) {
	msg := NewChatMessage(msgType, content)
	m.messages = append(m.messages, msg)
	m.updateViewport()
}

// ClearMessages clears all messages except the welcome message.
func (m *ChatModel) ClearMessages() {
	welcomeMsg := NewChatMessage(MessageTypeSystem, "🤖 **K9s Chat Assistant** ready! I can help you with Kubernetes operations.\n\nType your questions and I'll provide assistance based on your current k9s context.")
	m.messages = []*ChatMessage{welcomeMsg}
	m.updateViewport()
}

// SetError sets the error state.
func (m *ChatModel) SetError(err error) {
	m.err = err
}

// SetLoading sets the loading state.
func (m *ChatModel) SetLoading(loading bool) {
	m.loading = loading
}

// RenderHeader renders the chat header for external use.
func (m ChatModel) RenderHeader() string {
	return m.renderHeader()
}

// RenderMessages renders all messages for external use.
func (m ChatModel) RenderMessages() string {
	var content strings.Builder
	for i, msg := range m.messages {
		if i > 0 {
			content.WriteString("\n\n")
		}
		content.WriteString(m.renderMessage(msg))
	}
	return content.String()
}

// RenderInput renders the input area for external use.
func (m ChatModel) RenderInput() string {
	return m.renderInput()
}

// UpdateK9sContext updates the k9s context.
func (m *ChatModel) UpdateK9sContext(ctx *K9sContext) {
	m.k9sContext = ctx
}

// wordWrap wraps text to the specified width.
func wordWrap(text string, width int) string {
	if width <= 0 {
		return text
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return text
	}

	var lines []string
	var currentLine strings.Builder

	for _, word := range words {
		if currentLine.Len()+len(word)+1 > width {
			if currentLine.Len() > 0 {
				lines = append(lines, currentLine.String())
				currentLine.Reset()
			}
		}

		if currentLine.Len() > 0 {
			currentLine.WriteString(" ")
		}
		currentLine.WriteString(word)
	}

	if currentLine.Len() > 0 {
		lines = append(lines, currentLine.String())
	}

	return strings.Join(lines, "\n")
}
