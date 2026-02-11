// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"k8s.io/apimachinery/pkg/labels"
)

const (
	chatTitleFmt = "[fg:bg:b] kagent Chat ([hilite:bg:b]%s[fg:bg:-])[fg:bg:-] "
)

// KAgentChat represents an embedded chat view for kagent agents.
type KAgentChat struct {
	*tview.Flex

	app           *App
	agentName     string
	namespace     string
	input         *tview.InputField
	output        *tview.TextView
	actions       *ui.KeyActions
	history       []string
	historyIdx    int
}

// NewKAgentChat creates a new kagent chat view.
func NewKAgentChat(app *App, namespace, agentName string) *KAgentChat {
	c := &KAgentChat{
		Flex:       tview.NewFlex().SetDirection(tview.FlexRow),
		app:        app,
		agentName:  agentName,
		namespace:  namespace,
		output:     tview.NewTextView(),
		input:      tview.NewInputField(),
		actions:    ui.NewKeyActions(),
		history:    make([]string, 0),
		historyIdx: -1,
	}

	return c
}

func (*KAgentChat) SetCommand(*cmd.Interpreter)            {}
func (*KAgentChat) SetFilter(string, bool)                 {}
func (*KAgentChat) SetLabelSelector(labels.Selector, bool) {}

// Init initializes the chat view.
func (c *KAgentChat) Init(_ context.Context) error {
	c.SetBorder(true)
	c.SetBorderPadding(0, 0, 1, 1)
	c.updateTitle()

	// Configure output area
	c.output.SetScrollable(true)
	c.output.SetWrap(true)
	c.output.SetDynamicColors(true)
	c.output.SetChangedFunc(func() {
		c.app.Draw()
	})

	// Configure input field
	c.input.SetLabel("[green]> ")
	c.input.SetFieldBackgroundColor(tcell.ColorDefault)
	c.input.SetDoneFunc(c.handleInput)

	// Layout: output takes most space, input at bottom
	c.AddItem(c.output, 0, 1, false)
	c.AddItem(c.input, 1, 0, true)

	c.app.Styles.AddListener(c)
	c.StylesChanged(c.app.Styles)

	c.bindKeys()

	// Add welcome message
	c.appendOutput("[yellow]kagent Chat[white]\n")
	c.appendOutput(fmt.Sprintf("Connected to agent: [green]%s[white] in namespace [cyan]%s[white]\n", c.agentName, c.namespace))
	c.appendOutput("Type your message and press Enter to send.\n")
	c.appendOutput("Press [yellow]Ctrl+C[white] or [yellow]Esc[white] to exit.\n")
	c.appendOutput(strings.Repeat("-", 50) + "\n\n")

	return nil
}

func (c *KAgentChat) bindKeys() {
	c.actions.Bulk(ui.KeyMap{
		tcell.KeyEscape: ui.NewKeyAction("Back", c.backCmd, false),
		tcell.KeyCtrlC:  ui.NewKeyAction("Back", c.backCmd, false),
		tcell.KeyUp:     ui.NewKeyAction("History Up", c.historyUpCmd, false),
		tcell.KeyDown:   ui.NewKeyAction("History Down", c.historyDownCmd, false),
	})
}

func (c *KAgentChat) handleInput(key tcell.Key) {
	if key != tcell.KeyEnter {
		return
	}

	text := strings.TrimSpace(c.input.GetText())
	if text == "" {
		return
	}

	// Add to history
	c.history = append(c.history, text)
	c.historyIdx = len(c.history)

	// Clear input
	c.input.SetText("")

	// Show user message
	c.appendOutput(fmt.Sprintf("[green]You:[white] %s\n\n", text))

	// Send to agent
	go c.sendMessage(text)
}

func (c *KAgentChat) sendMessage(message string) {
	c.appendOutput("[yellow]Agent thinking...[white]\n")

	// Check if kagent CLI exists
	kagentPath, err := exec.LookPath("kagent")
	if err != nil {
		c.appendOutput("[red]Error:[white] kagent CLI not found. Install from: https://kagent.dev\n\n")
		return
	}

	// Build command with JSON output
	args := []string{
		"invoke",
		"--namespace", c.namespace,
		"--agent", c.agentName,
		"--task", message,
		"--output-format", "json",
	}

	cmd := exec.Command(kagentPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		c.appendOutput(fmt.Sprintf("[red]Error:[white] %v\n", err))
	}

	// Parse JSON response and extract text
	response := c.parseAgentResponse(string(output))
	if response != "" {
		c.appendOutput(fmt.Sprintf("[cyan]Agent:[white]\n%s\n\n", response))
	}
}

// parseAgentResponse extracts text from kagent JSON response
func (c *KAgentChat) parseAgentResponse(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}

	// Try to parse as JSON
	var resp struct {
		Artifacts []struct {
			Parts []struct {
				Kind string `json:"kind"`
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"artifacts"`
	}

	if err := json.Unmarshal([]byte(raw), &resp); err != nil {
		// Not JSON, return as-is
		return raw
	}

	// Extract text from artifacts
	var texts []string
	for _, artifact := range resp.Artifacts {
		for _, part := range artifact.Parts {
			if part.Kind == "text" && part.Text != "" {
				texts = append(texts, part.Text)
			}
		}
	}

	if len(texts) > 0 {
		return strings.Join(texts, "\n")
	}

	return raw
}

func (c *KAgentChat) appendOutput(text string) {
	c.app.QueueUpdateDraw(func() {
		fmt.Fprintf(c.output, "%s", text)
		c.output.ScrollToEnd()
	})
}

func (c *KAgentChat) backCmd(*tcell.EventKey) *tcell.EventKey {
	return c.app.PrevCmd(nil)
}

func (c *KAgentChat) historyUpCmd(*tcell.EventKey) *tcell.EventKey {
	if len(c.history) == 0 {
		return nil
	}
	if c.historyIdx > 0 {
		c.historyIdx--
	}
	c.input.SetText(c.history[c.historyIdx])
	return nil
}

func (c *KAgentChat) historyDownCmd(*tcell.EventKey) *tcell.EventKey {
	if len(c.history) == 0 {
		return nil
	}
	if c.historyIdx < len(c.history)-1 {
		c.historyIdx++
		c.input.SetText(c.history[c.historyIdx])
	} else {
		c.historyIdx = len(c.history)
		c.input.SetText("")
	}
	return nil
}

func (c *KAgentChat) updateTitle() {
	title := fmt.Sprintf(chatTitleFmt, c.agentName)
	styles := c.app.Styles.Frame()
	c.SetTitle(ui.SkinTitle(title, &styles))
}

// StylesChanged handles style changes.
func (c *KAgentChat) StylesChanged(s *config.Styles) {
	c.SetBackgroundColor(s.BgColor())
	c.output.SetTextColor(s.FgColor())
	c.output.SetBackgroundColor(s.BgColor())
	c.input.SetFieldTextColor(s.FgColor())
}

// InCmdMode checks if in command mode.
func (*KAgentChat) InCmdMode() bool { return false }

// Name returns the view name.
func (c *KAgentChat) Name() string { return "kagent-chat" }

// Start starts the view.
func (*KAgentChat) Start() {}

// Stop stops the view.
func (c *KAgentChat) Stop() {
	c.app.Styles.RemoveListener(c)
}

// Hints returns menu hints.
func (c *KAgentChat) Hints() model.MenuHints {
	return c.actions.Hints()
}

// ExtraHints returns extra hints.
func (*KAgentChat) ExtraHints() map[string]string {
	return nil
}

// Actions returns key actions.
func (c *KAgentChat) Actions() *ui.KeyActions {
	return c.actions
}
