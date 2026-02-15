// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/ai"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	claudeTitle    = "Claude"
	claudeTitleFmt = " [aqua::b]%s "
)

// Claude represents the Claude AI chat view.
type Claude struct {
	*tview.Flex

	app         *App
	actions     *ui.KeyActions
	cmdBuff     *model.FishBuff
	chatHistory *tview.TextView
	contextInfo *tview.TextView
	messages    []ai.Message
	k8sContext  *ai.K8sContext
}

// NewClaude returns a new Claude view instance.
func NewClaude(app *App, question string) *Claude {
	c := &Claude{
		Flex:        tview.NewFlex(),
		app:         app,
		actions:     ui.NewKeyActions(),
		cmdBuff:     model.NewFishBuff(':', model.CommandBuffer),
		chatHistory: tview.NewTextView(),
		contextInfo: tview.NewTextView(),
		messages:    make([]ai.Message, 0),
	}

	c.buildK8sContext()

	if question != "" {
		c.messages = append(c.messages, ai.Message{
			Role:    "user",
			Content: question,
		})
	}

	return c
}

func (*Claude) SetCommand(*cmd.Interpreter)            {}
func (*Claude) SetFilter(string, bool)                 {}
func (*Claude) SetLabelSelector(labels.Selector, bool) {}

// Init initializes the view.
func (c *Claude) Init(_ context.Context) error {
	c.SetDirection(tview.FlexRow)
	c.SetBorder(true)
	c.SetTitle(fmt.Sprintf(claudeTitleFmt, claudeTitle))
	c.SetTitleColor(tcell.ColorAqua)
	c.SetBorderPadding(0, 0, 1, 1)

	// Context info section
	c.contextInfo.SetDynamicColors(true)
	c.contextInfo.SetBorder(true)
	c.contextInfo.SetTitle(" Context ")
	c.contextInfo.SetBorderColor(tcell.ColorGray)
	c.updateContextDisplay()

	// Chat history section
	c.chatHistory.SetDynamicColors(true)
	c.chatHistory.SetScrollable(true)
	c.chatHistory.SetWrap(true)
	c.chatHistory.SetBorder(true)
	c.chatHistory.SetTitle(" Chat ")
	c.chatHistory.SetBorderColor(tcell.ColorGray)

	// Layout
	c.AddItem(c.contextInfo, 5, 0, false)
	c.AddItem(c.chatHistory, 0, 1, true)

	c.app.Styles.AddListener(c)
	c.StylesChanged(c.app.Styles)

	c.bindKeys()
	c.SetInputCapture(c.keyboard)

	c.app.Prompt().SetModel(c.cmdBuff)
	c.cmdBuff.AddListener(c)

	// If we have an initial question, send it
	if len(c.messages) > 0 {
		go c.sendMessage()
	}

	return nil
}

func (c *Claude) buildK8sContext() {
	c.k8sContext = &ai.K8sContext{}

	if c.app.Conn() != nil && c.app.Conn().ConnectionOK() {
		cfg := c.app.Conn().Config()
		if cfg != nil {
			c.k8sContext.ContextName = c.app.Config.ActiveContextName()
			if clusterName, err := cfg.CurrentClusterName(); err == nil {
				c.k8sContext.ClusterName = clusterName
			}
		}
	}

	c.k8sContext.Namespace = c.app.Config.ActiveNamespace()

	// Get current view info
	if top := c.app.Content.Top(); top != nil {
		c.k8sContext.ResourceType = top.Name()
	}

	// Try to get selected resource info from the current view
	c.extractSelectedResource()
}

func (c *Claude) extractSelectedResource() {
	top := c.app.Content.Top()
	if top == nil {
		return
	}

	// Try to get ResourceViewer interface
	if rv, ok := top.(ResourceViewer); ok {
		if tbl := rv.GetTable(); tbl != nil {
			if sel := tbl.GetSelectedItem(); sel != "" {
				c.k8sContext.SelectedResource = sel
			}
		}
	}
}

func (c *Claude) updateContextDisplay() {
	var sb strings.Builder
	sb.WriteString("[yellow]Cluster:[white] ")
	if c.k8sContext.ClusterName != "" {
		sb.WriteString(c.k8sContext.ClusterName)
	} else {
		sb.WriteString("N/A")
	}
	sb.WriteString("  [yellow]Context:[white] ")
	if c.k8sContext.ContextName != "" {
		sb.WriteString(c.k8sContext.ContextName)
	} else {
		sb.WriteString("N/A")
	}
	sb.WriteString("  [yellow]Namespace:[white] ")
	if c.k8sContext.Namespace != "" {
		sb.WriteString(c.k8sContext.Namespace)
	} else {
		sb.WriteString("all")
	}
	if c.k8sContext.ResourceType != "" {
		sb.WriteString("  [yellow]View:[white] ")
		sb.WriteString(c.k8sContext.ResourceType)
	}
	if c.k8sContext.SelectedResource != "" {
		sb.WriteString("\n[yellow]Selected:[white] ")
		sb.WriteString(c.k8sContext.SelectedResource)
	}

	c.contextInfo.SetText(sb.String())
}

func (c *Claude) updateChatDisplay() {
	var sb strings.Builder

	for _, msg := range c.messages {
		switch msg.Role {
		case "user":
			sb.WriteString("[aqua::b]You:[white:-:-] ")
			sb.WriteString(msg.Content)
			sb.WriteString("\n\n")
		case "assistant":
			sb.WriteString("[green::b]Claude:[white:-:-] ")
			sb.WriteString(msg.Content)
			sb.WriteString("\n\n")
		}
	}

	c.chatHistory.SetText(sb.String())
	c.chatHistory.ScrollToEnd()
}

func (c *Claude) sendMessage() {
	apiKey := c.app.Config.K9s.AI.GetAPIKey()
	if apiKey == "" {
		c.app.QueueUpdateDraw(func() {
			c.messages = append(c.messages, ai.Message{
				Role:    "assistant",
				Content: "[red]Error: API key not configured. Use ':claude set-key <your-api-key>' to set it.",
			})
			c.updateChatDisplay()
		})
		return
	}

	// Show loading indicator
	c.app.QueueUpdateDraw(func() {
		c.chatHistory.SetText(c.chatHistory.GetText(false) + "[gray]Thinking...[white]\n")
	})

	client := ai.NewClient(
		apiKey,
		c.app.Config.K9s.AI.GetModel(),
		c.app.Config.K9s.AI.GetMaxTokens(),
	)

	systemPrompt, err := ai.BuildSystemPrompt(c.k8sContext)
	if err != nil {
		c.app.QueueUpdateDraw(func() {
			c.messages = append(c.messages, ai.Message{
				Role:    "assistant",
				Content: fmt.Sprintf("[red]Error building prompt: %v", err),
			})
			c.updateChatDisplay()
		})
		return
	}

	resp, err := client.Send(systemPrompt, c.messages)
	if err != nil {
		c.app.QueueUpdateDraw(func() {
			c.messages = append(c.messages, ai.Message{
				Role:    "assistant",
				Content: fmt.Sprintf("[red]Error: %v", err),
			})
			c.updateChatDisplay()
		})
		return
	}

	c.app.QueueUpdateDraw(func() {
		c.messages = append(c.messages, ai.Message{
			Role:    "assistant",
			Content: resp.GetText(),
		})
		c.updateChatDisplay()
	})
}

func (c *Claude) bindKeys() {
	c.actions.Bulk(ui.KeyMap{
		tcell.KeyEscape: ui.NewKeyAction("Back", c.backCmd, true),
		ui.KeyQ:         ui.NewKeyAction("Back", c.backCmd, false),
		tcell.KeyEnter:  ui.NewKeyAction("Send", c.sendCmd, true),
		tcell.KeyCtrlL:  ui.NewKeyAction("Clear", c.clearCmd, true),
		ui.KeyColon:     ui.NewSharedKeyAction("Prompt", c.activateCmd, false),
	})
}

func (c *Claude) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	if a, ok := c.actions.Get(ui.AsKey(evt)); ok {
		return a.Action(evt)
	}

	return evt
}

func (c *Claude) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if c.cmdBuff.InCmdMode() {
		c.cmdBuff.SetActive(false)
		c.cmdBuff.Reset()
		return nil
	}
	return c.app.PrevCmd(evt)
}

func (c *Claude) sendCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !c.cmdBuff.InCmdMode() || c.cmdBuff.Empty() {
		return evt
	}

	question := c.cmdBuff.GetText()
	c.cmdBuff.SetActive(false)
	c.cmdBuff.Reset()

	c.messages = append(c.messages, ai.Message{
		Role:    "user",
		Content: question,
	})
	c.updateChatDisplay()

	go c.sendMessage()

	return nil
}

func (c *Claude) clearCmd(*tcell.EventKey) *tcell.EventKey {
	c.messages = make([]ai.Message, 0)
	c.chatHistory.SetText("")
	return nil
}

func (c *Claude) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if c.app.InCmdMode() {
		return evt
	}
	c.app.ResetPrompt(c.cmdBuff)
	return nil
}

// BufferChanged indicates the buffer was changed.
func (*Claude) BufferChanged(_, _ string) {}

// BufferCompleted indicates input was accepted.
func (c *Claude) BufferCompleted(text, _ string) {
	if text == "" {
		return
	}

	c.messages = append(c.messages, ai.Message{
		Role:    "user",
		Content: text,
	})
	c.updateChatDisplay()

	go c.sendMessage()
}

// BufferActive indicates the buff activity changed.
func (c *Claude) BufferActive(state bool, k model.BufferKind) {
	c.app.BufferActive(state, k)
}

// InCmdMode checks if prompt is active.
func (c *Claude) InCmdMode() bool {
	return c.cmdBuff.InCmdMode()
}

// StylesChanged notifies the skin changed.
func (c *Claude) StylesChanged(s *config.Styles) {
	c.SetBackgroundColor(s.BgColor())
	c.chatHistory.SetBackgroundColor(s.BgColor())
	c.chatHistory.SetTextColor(s.FgColor())
	c.contextInfo.SetBackgroundColor(s.BgColor())
	c.contextInfo.SetTextColor(s.FgColor())
	c.SetBorderFocusColor(s.Frame().Border.FocusColor.Color())
}

// Name returns the component name.
func (*Claude) Name() string { return claudeTitle }

// Start starts the view updater.
func (*Claude) Start() {}

// Stop terminates the updater.
func (c *Claude) Stop() {
	c.app.Styles.RemoveListener(c)
}

// Hints returns menu hints.
func (c *Claude) Hints() model.MenuHints {
	return c.actions.Hints()
}

// ExtraHints returns additional hints.
func (*Claude) ExtraHints() map[string]string {
	return nil
}

// App returns the app reference.
func (c *Claude) App() *App {
	return c.app
}

// GetContext returns the context from app.
func (c *Claude) GetContext() context.Context {
	return context.WithValue(context.Background(), internal.KeyApp, c.app)
}
