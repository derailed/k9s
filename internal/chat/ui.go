// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package chat

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/chat/model"
	"github.com/derailed/k9s/internal/chat/provider"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// ChatUI represents the chat user interface.
type ChatUI struct {
	*tview.Flex
	textView   *tview.TextView
	inputView  *tview.InputField
	chatModel  *model.ChatModel
	provider   provider.LLMProvider
	k9sContext *model.K9sContext
	focused    bool
	sending    bool
}

// NewChatUI creates a new chat UI.
func NewChatUI(p provider.LLMProvider, ctx *model.K9sContext) *ChatUI {
	// Create text view for messages
	textView := tview.NewTextView()
	textView.SetDynamicColors(true)
	textView.SetRegions(true)
	textView.SetScrollable(true)
	textView.SetWrap(true)
	textView.SetWordWrap(true)
	textView.SetBorder(false)

	// Create input field
	inputView := tview.NewInputField()
	inputView.SetLabel("❯ ")
	inputView.SetPlaceholder("Ask me about Kubernetes...")
	inputView.SetFieldWidth(0)
	inputView.SetBorder(true)
	inputView.SetTitle(" Input ")
	inputView.SetTitleAlign(tview.AlignLeft)

	// Create flex container
	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow)
	flex.AddItem(textView, 0, 1, false) // Messages take all available space
	flex.AddItem(inputView, 3, 0, true) // Input is fixed 3 lines
	flex.SetBorder(true)
	flex.SetTitle(" 🤖 K9s Chat Assistant ")
	flex.SetTitleAlign(tview.AlignLeft)

	chatModel := model.NewChatModel(p, ctx)

	ui := &ChatUI{
		Flex:       flex,
		textView:   textView,
		inputView:  inputView,
		chatModel:  chatModel,
		provider:   p,
		k9sContext: ctx,
	}

	// Set up input handling
	inputView.SetDoneFunc(ui.handleSubmit)
	inputView.SetInputCapture(ui.handleInputKey)

	// Initialize with welcome message
	ui.refreshDisplay()

	return ui
}

// Start starts the chat UI.
func (ui *ChatUI) Start() {
	ui.focused = true
	ui.refreshDisplay()
	slog.Debug("Starting chat UI", slogs.Component, "chat")
}

// Stop stops the chat UI.
func (ui *ChatUI) Stop() {
	ui.focused = false
	slog.Debug("Stopping chat UI", slogs.Component, "chat")
}

// HandleKey handles key events for the chat UI.
func (ui *ChatUI) HandleKey(event *tcell.EventKey) *tcell.EventKey {
	if !ui.focused {
		return event
	}

	// Handle special keys when chat is focused
	switch event.Key() {
	case tcell.KeyTab:
		// Tab should not be handled here - let the parent handle focus switching
		return event
	case tcell.KeyEscape:
		// Clear input
		ui.inputView.SetText("")
		return nil
	case tcell.KeyCtrlC:
		// Clear input
		ui.inputView.SetText("")
		return nil
	case tcell.KeyCtrlL:
		// Clear chat
		ui.ClearChat()
		return nil
	}

	// For other keys, let the input field handle them
	return event
}

// Focus sets focus to the chat UI.
func (ui *ChatUI) Focus(delegate func(p tview.Primitive)) {
	ui.focused = true
	delegate(ui.inputView)
}

// Blur removes focus from the chat UI.
func (ui *ChatUI) Blur() {
	ui.focused = false
	ui.Flex.Blur()
}

// HasFocus returns whether the chat UI has focus.
func (ui *ChatUI) HasFocus() bool {
	return ui.focused
}

// GetInputField returns the input field for external access.
func (ui *ChatUI) GetInputField() *tview.InputField {
	return ui.inputView
}

// UpdateContext updates the k9s context.
func (ui *ChatUI) UpdateContext(ctx *model.K9sContext) {
	ui.k9sContext = ctx
	ui.chatModel.UpdateK9sContext(ctx)
	ui.refreshDisplay()
}

// refreshDisplay updates the entire chat display.
func (ui *ChatUI) refreshDisplay() {
	header := ui.chatModel.RenderHeader()
	messages := ui.chatModel.RenderMessages()

	// Combine header and messages
	content := header
	if messages != "" {
		content += "\n\n" + messages
	}

	// Add loading indicator if sending
	if ui.sending {
		content += "\n\n[yellow]🤖 Assistant:[white] [blink]◐ Thinking...[white]"
	}

	ui.textView.SetText(content)
	ui.textView.ScrollToEnd()
}

// handleSubmit handles message submission.
func (ui *ChatUI) handleSubmit(key tcell.Key) {
	if key == tcell.KeyEnter {
		text := strings.TrimSpace(ui.inputView.GetText())
		if text != "" {
			ui.SendMessage(text)
			ui.inputView.SetText("")
		}
	}
}

// handleInputKey handles input field key events.
func (ui *ChatUI) handleInputKey(event *tcell.EventKey) *tcell.EventKey {
	switch event.Key() {
	case tcell.KeyEscape, tcell.KeyCtrlU:
		ui.inputView.SetText("")
		return nil
	}
	return event
}

// SendMessage sends a user message and gets LLM response.
func (ui *ChatUI) SendMessage(text string) {
	if ui.sending {
		return // Prevent multiple concurrent requests
	}

	// Add user message
	ui.chatModel.AddMessage(model.MessageTypeUser, text)
	ui.sending = true
	ui.refreshDisplay()

	// Send to LLM in background
	go ui.fetchLLMResponse(text)
}

// fetchLLMResponse gets response from LLM provider.
func (ui *ChatUI) fetchLLMResponse(userMessage string) {
	defer func() {
		ui.sending = false
		ui.refreshDisplay()
	}()

	// Prepare messages including system context
	messages := []provider.Message{
		{Role: "system", Content: ui.k9sContext.GetSystemContext()},
		{Role: "user", Content: userMessage},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := ui.provider.SendMessage(ctx, messages, nil)
	if err != nil {
		ui.chatModel.SetError(err)
		slog.Error("LLM response failed", slogs.Error, err)
		return
	}

	ui.chatModel.AddMessage(model.MessageTypeAssistant, resp.Content)
	ui.chatModel.SetError(nil)
}

// ClearChat clears all messages.
func (ui *ChatUI) ClearChat() {
	ui.chatModel.ClearMessages()
	ui.refreshDisplay()
}

// ClearInput clears the input field (for external use).
func (ui *ChatUI) ClearInput() {
	ui.inputView.SetText("")
}
