// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package chat

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"k8s.io/apimachinery/pkg/labels"
)

// Component represents the main chat component.
type Component struct {
	*tview.Flex

	chatView *ChatView
	app      App
	actions  *ui.KeyActions
	styles   *config.Styles
	visible  bool
}

// NewChatComponent returns a new chat component (for k9s integration).
func NewChatComponent(application *tview.Application, factory interface{}) *Component {
	// Create a wrapper app that implements our App interface
	app := &appWrapper{application: application, factory: factory}
	return NewComponent(app)
}

// NewComponent returns a new chat component.
func NewComponent(app App) *Component {
	c := &Component{
		Flex:    tview.NewFlex(),
		app:     app,
		actions: ui.NewKeyActions(),
		visible: false,
	}

	c.chatView = NewChatView(app)
	c.SetDirection(tview.FlexRow)
	c.AddItem(c.chatView, 0, 1, true)
	c.bindKeys()

	return c
}

// Init initializes the component.
func (c *Component) Init(ctx context.Context) error {
	if err := c.chatView.Init(ctx); err != nil {
		return err
	}

	c.SetBorder(true)
	c.SetTitle(" Chat Assistant ")
	c.SetBorderPadding(0, 0, 1, 1)

	return nil
}

// Name returns the component name.
func (c *Component) Name() string {
	return "chat"
}

// Start starts the component.
func (c *Component) Start() {
	c.visible = true
}

// Stop stops the component.
func (c *Component) Stop() {
	c.visible = false
}

// Hints returns the component hints.
func (c *Component) Hints() model.MenuHints {
	return c.actions.Hints()
}

// ExtraHints returns additional hints.
func (c *Component) ExtraHints() map[string]string {
	return nil
}

// InCmdMode checks if the component is in command mode.
func (c *Component) InCmdMode() bool {
	return false
}

// SetCommand sets the current command (not used for chat).
func (c *Component) SetCommand(interpreter *cmd.Interpreter) {
	// Chat doesn't use commands in the same way as other views
}

// SetFilter sets the filter (not used for chat).
func (c *Component) SetFilter(filter string) {
	// Chat doesn't use filters
}

// SetLabelSelector sets the label selector (not used for chat).
func (c *Component) SetLabelSelector(selector labels.Selector) {
	// Chat doesn't use label selectors
}

// StylesChanged notifies the component that styles have changed.
func (c *Component) StylesChanged(s *config.Styles) {
	c.styles = s
	c.SetBackgroundColor(s.BgColor())
	c.SetBorderColor(s.Frame().Border.FgColor.Color())
	c.SetTitleColor(s.Frame().Title.FgColor.Color())

	c.chatView.StylesChanged(s)
}

// Focus gives focus to the component.
func (c *Component) Focus(delegate func(p tview.Primitive)) {
	delegate(c.chatView)
}

// HasFocus returns true if the component has focus.
func (c *Component) HasFocus() bool {
	return c.chatView.HasFocus()
}

func (c *Component) bindKeys() {
	c.actions.Bulk(ui.KeyMap{
		// Chat-specific actions that will appear in k9s header menu
		tcell.KeyUp:    ui.NewKeyAction("Browse Messages", c.noopCmd, true),
		tcell.KeyDown:  ui.NewKeyAction("Browse Messages", c.noopCmd, true),
		tcell.KeyEnter: ui.NewKeyAction("Type Message", c.noopCmd, true),
		tcell.KeyCtrlL: ui.NewKeyAction("Clear Chat", c.clearCmd, true),
		tcell.KeyTab:   ui.NewKeyAction("Switch Focus", c.switchFocusCmd, true),
	})
}

func (c *Component) noopCmd(evt *tcell.EventKey) *tcell.EventKey {
	// These actions are handled by the chat view itself
	return evt
}

func (c *Component) clearCmd(evt *tcell.EventKey) *tcell.EventKey {
	// Clear chat messages
	if c.chatView != nil {
		c.chatView.clearChat(evt)
	}
	return nil
}

func (c *Component) switchFocusCmd(evt *tcell.EventKey) *tcell.EventKey {
	// Switch focus between chat and main k9s view
	if c.app != nil {
		c.app.SwitchFocus()
	}
	return nil
}

// SendWelcomeMessage sends an initial welcome message.
func (c *Component) SendWelcomeMessage() {
	welcomeMsg := fmt.Sprintf(`# Welcome to K9s Chat Assistant! 🚀

I'm here to help you with Kubernetes operations. I can:

- **Answer questions** about your current cluster and resources
- **Suggest kubectl commands** based on your current context
- **Explain Kubernetes concepts** and best practices
- **Help troubleshoot** issues you're experiencing

**Current Context:**
- Cluster: %s
- Namespace: %s
- Current View: %s

Type your question below and press **Enter** to send!

*Use **Tab** to switch focus between chat and main view*  
*Use :chat command to close chat*`,
		c.getClusterName(),
		c.getCurrentNamespace(),
		c.getCurrentView())

	c.chatView.AddBotMessage(welcomeMsg)
}

func (c *Component) getClusterName() string {
	if c.app == nil {
		return "unknown"
	}
	return c.app.GetClusterName()
}

func (c *Component) getCurrentNamespace() string {
	if c.app == nil {
		return "default"
	}
	return c.app.GetCurrentNamespace()
}

func (c *Component) getCurrentView() string {
	if c.app == nil {
		return "unknown"
	}
	return c.app.GetCurrentView()
}
