// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package chat

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// ChatView represents the main chat interface.
type ChatView struct {
	*tview.Flex

	messageView *MarkdownTextView
	inputField  *tview.InputField
	statusBar   *tview.TextView
	actions     *ui.KeyActions
	styles      *config.Styles
	app         App
	provider    Provider

	messages []Message
	focused  string // "input" or "messages"
}

// NewChatView returns a new chat view.
func NewChatView(app App) *ChatView {
	cv := &ChatView{
		Flex:     tview.NewFlex(),
		app:      app,
		actions:  ui.NewKeyActions(),
		messages: make([]Message, 0),
		focused:  "input",
	}

	cv.setupComponents()
	cv.setupLayout()
	cv.bindKeys()

	// Initialize mock provider
	cv.provider = NewMockProvider(app)

	return cv
}

func (cv *ChatView) setupComponents() {
	// Message view (scrollable, read-only) - using our custom MarkdownTextView
	cv.messageView = NewMarkdownTextView()
	cv.messageView.SetBorder(true)
	cv.messageView.SetTitle(" Messages ")

	// Input field
	cv.inputField = tview.NewInputField()
	cv.inputField.SetLabel("You: ")
	cv.inputField.SetFieldWidth(0) // Use all available width
	cv.inputField.SetBorder(true)
	cv.inputField.SetTitle(" Type your message (Enter to send) ")

	// Status bar
	cv.statusBar = tview.NewTextView()
	cv.statusBar.SetDynamicColors(true)
	cv.statusBar.SetText("Ready • Tab: switch focus • Ctrl+L: clear")
	cv.statusBar.SetTextAlign(tview.AlignCenter)
}

func (cv *ChatView) setupLayout() {
	cv.SetDirection(tview.FlexRow)

	// Full-screen optimized layout with more space for messages
	// Message view (takes most space) - optimized for full-screen
	cv.AddItem(cv.messageView, 0, 10, false)

	// Input field - slightly larger for better visibility
	cv.AddItem(cv.inputField, 4, 1, true)

	// Status bar
	cv.AddItem(cv.statusBar, 1, 1, false)
}

// Init initializes the chat view.
func (cv *ChatView) Init(ctx context.Context) error {
	cv.inputField.SetInputCapture(cv.inputCapture)
	cv.messageView.SetInputCapture(cv.messageCapture)

	return nil
}

// StylesChanged updates the component styles.
func (cv *ChatView) StylesChanged(s *config.Styles) {
	cv.styles = s

	// Apply styles to components
	bgColor := s.BgColor()
	borderColor := s.Frame().Border.FgColor.Color()
	titleColor := s.Frame().Title.FgColor.Color()

	cv.SetBackgroundColor(bgColor)

	cv.messageView.SetBackgroundColor(bgColor)
	cv.messageView.SetBorderColor(borderColor)
	cv.messageView.SetTitleColor(titleColor)

	cv.inputField.SetBackgroundColor(bgColor)
	cv.inputField.SetBorderColor(borderColor)
	cv.inputField.SetTitleColor(titleColor)
	cv.inputField.SetFieldBackgroundColor(bgColor)

	cv.statusBar.SetBackgroundColor(bgColor)
	cv.statusBar.SetTextColor(s.FgColor())
}

func (cv *ChatView) bindKeys() {
	cv.actions.Bulk(ui.KeyMap{
		tcell.KeyEnter: ui.NewKeyAction("Send Message", cv.sendMessage, true),
		tcell.KeyTab:   ui.NewKeyAction("Switch Focus", cv.switchFocus, true),
		tcell.KeyCtrlL: ui.NewKeyAction("Clear Chat", cv.clearChat, true),
	})
}

func (cv *ChatView) inputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEnter:
		cv.sendMessage(event)
		return nil
	case tcell.KeyTab:
		cv.switchFocus(event)
		return nil
	case tcell.KeyCtrlL:
		cv.clearChat(event)
		return nil
	case tcell.KeyUp, tcell.KeyDown:
		// When typing, up/down should switch to message browsing
		if cv.focused == "input" {
			cv.focused = "messages"
			cv.app.SetFocus(cv.messageView)
			cv.statusBar.SetText("📖 Browsing messages • ⇥: to input • ⏎: type • ↑↓: scroll")
			// Let the message view handle the up/down
			return event
		}
	}
	return event
}

func (cv *ChatView) messageCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEnter:
		// Enter while browsing = go back to input to type
		cv.focused = "input"
		cv.app.SetFocus(cv.inputField)
		cv.statusBar.SetText("💬 Typing mode • ⇥: to messages • ↑↓: browse • ⏎: send")
		return nil
	case tcell.KeyTab:
		cv.switchFocus(event)
		return nil
	case tcell.KeyCtrlL:
		cv.clearChat(event)
		return nil
	}
	return event
}

// Focus management
func (cv *ChatView) Focus(delegate func(p tview.Primitive)) {
	if cv.focused == "input" {
		delegate(cv.inputField)
	} else {
		delegate(cv.messageView)
	}
}

func (cv *ChatView) HasFocus() bool {
	return cv.inputField.HasFocus() || cv.messageView.HasFocus()
}

// Message handling
func (cv *ChatView) AddUserMessage(content string) {
	msg := Message{
		ID:        fmt.Sprintf("user_%d", time.Now().UnixNano()),
		Content:   content,
		Type:      MessageTypeUser,
		Timestamp: time.Now(),
	}
	cv.messages = append(cv.messages, msg)
	cv.renderMessage(msg)
}

func (cv *ChatView) AddBotMessage(content string) {
	msg := Message{
		ID:        fmt.Sprintf("bot_%d", time.Now().UnixNano()),
		Content:   content,
		Type:      MessageTypeAssistant,
		Timestamp: time.Now(),
	}
	cv.messages = append(cv.messages, msg)
	cv.renderMessage(msg)
}

func (cv *ChatView) renderMessage(msg Message) {
	timestamp := msg.Timestamp.Format("15:04")

	// Create a complete markdown message including header
	var messageMarkdown string
	switch msg.Type {
	case MessageTypeUser:
		messageMarkdown = fmt.Sprintf("**👤 You** _%s_\n\n%s\n\n", timestamp, msg.Content)
	case MessageTypeAssistant:
		messageMarkdown = fmt.Sprintf("**🤖 Assistant** _%s_\n\n%s\n\n", timestamp, msg.Content)
	case MessageTypeSystem:
		messageMarkdown = fmt.Sprintf("**ℹ️ System** _%s_\n\n%s\n\n", timestamp, msg.Content)
	case MessageTypeError:
		messageMarkdown = fmt.Sprintf("**❌ Error** _%s_\n\n%s\n\n", timestamp, msg.Content)
	}

	// Render everything as markdown for consistent, beautiful formatting
	cv.messageView.AddMarkdown(messageMarkdown)

	// Scroll to bottom
	cv.messageView.ScrollToEnd()
}

// Command handlers
func (cv *ChatView) sendMessage(evt *tcell.EventKey) *tcell.EventKey {
	text := strings.TrimSpace(cv.inputField.GetText())
	if text == "" {
		return nil
	}

	// Add user message
	cv.AddUserMessage(text)

	// Clear input
	cv.inputField.SetText("")

	// Update status
	cv.statusBar.SetText("🤔 Thinking...")

	// Get response from provider (async)
	go func() {
		response, err := cv.provider.GetResponse(text, cv.getK9sContext())

		cv.app.QueueUpdateDraw(func() {
			if err != nil {
				cv.AddBotMessage(fmt.Sprintf("Sorry, I encountered an error: %v", err))
			} else {
				cv.AddBotMessage(response)
			}
			cv.statusBar.SetText("Ready • Tab: switch focus • Ctrl+L: clear")
		})
	}()

	return nil
}

func (cv *ChatView) switchFocus(evt *tcell.EventKey) *tcell.EventKey {
	if cv.focused == "input" {
		cv.focused = "messages"
		cv.app.SetFocus(cv.messageView)
		cv.statusBar.SetText("📖 Browsing messages • ⇥: to input • ⏎: type • ↑↓: scroll")
	} else {
		cv.focused = "input"
		cv.app.SetFocus(cv.inputField)
		cv.statusBar.SetText("💬 Typing mode • ⇥: to messages • ↑↓: browse • ⏎: send")
	}
	return nil
}

func (cv *ChatView) clearChat(evt *tcell.EventKey) *tcell.EventKey {
	cv.messages = make([]Message, 0)
	cv.messageView.Clear()
	cv.statusBar.SetText("Chat cleared! Ready for new conversation")
	return nil
}

func (cv *ChatView) getK9sContext() K9sContext {
	ctx := K9sContext{
		CurrentNamespace: "default",
		CurrentView:      "unknown",
		ClusterName:      "default",
	}

	if cv.app != nil {
		ctx.CurrentNamespace = cv.app.GetCurrentNamespace()
		ctx.CurrentView = cv.app.GetCurrentView()
		ctx.ClusterName = cv.app.GetClusterName()
	}

	return ctx
}
