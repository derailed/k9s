package ui

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

const defaultPrompt = "%c> [::b]%s"

// Command captures users free from command input.
type Command struct {
	*tview.TextView

	activated       bool
	icon            rune
	text            string
	styles          *config.Styles
	model           *model.FishBuff
	suggestions     []string
	suggestionIndex int
}

// NewCommand returns a new command view.
func NewCommand(styles *config.Styles, m *model.FishBuff) *Command {
	c := Command{styles: styles, TextView: tview.NewTextView(), model: m}
	c.SetWordWrap(true)
	c.SetWrap(true)
	c.SetDynamicColors(true)
	c.SetBorder(true)
	c.SetBorderPadding(0, 0, 1, 1)
	c.SetBackgroundColor(styles.BgColor())
	c.SetTextColor(styles.FgColor())
	styles.AddListener(&c)
	c.SetInputCapture(c.keyboard)

	return &c
}

func (c *Command) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	switch evt.Key() {
	case tcell.KeyEnter, tcell.KeyCtrlE:
		if c.suggestionIndex >= 0 {
			c.model.Set(c.text + c.suggestions[c.suggestionIndex])
		}
	case tcell.KeyCtrlW, tcell.KeyCtrlU:
		c.model.Clear()
	case tcell.KeyDown:
		if c.text == "" || c.suggestionIndex < 0 {
			return evt
		}
		c.suggestionIndex++
		if c.suggestionIndex >= len(c.suggestions) {
			c.suggestionIndex = 0
		}
		c.suggest(c.model.String(), c.suggestions[c.suggestionIndex])
	case tcell.KeyUp:
		if c.text == "" || c.suggestionIndex < 0 {
			return evt
		}
		c.suggestionIndex--
		if c.suggestionIndex < 0 {
			c.suggestionIndex = len(c.suggestions) - 1
		}
		c.suggest(c.model.String(), c.suggestions[c.suggestionIndex])
	case tcell.KeyTab, tcell.KeyRight, tcell.KeyCtrlF:
		if c.suggestionIndex >= 0 {
			c.model.Set(c.model.String() + c.suggestions[c.suggestionIndex])
			c.suggestionIndex = -1
		}
	}
	return evt
}

// StylesChanged notifies skin changed.
func (c *Command) StylesChanged(s *config.Styles) {
	c.styles = s
	c.SetBackgroundColor(s.BgColor())
	c.SetTextColor(s.FgColor())
}

// InCmdMode returns true if command is active, false otherwise.
func (c *Command) InCmdMode() bool {
	return c.activated
}

func (c *Command) activate() {
	c.write(c.text)
}

func (c *Command) update(s string) {
	if c.text == s {
		return
	}
	c.text = s
	c.Clear()
	c.write(c.text)
}

func (c *Command) suggest(text, suggestion string) {
	c.Clear()
	c.write(text + "[gray::-]" + suggestion)
}

func (c *Command) write(s string) {
	fmt.Fprintf(c, defaultPrompt, c.icon, s)
}

// ----------------------------------------------------------------------------
// Event Listener protocol...

// SuggestionChanged indicates the suggestions changed.
func (c *Command) SuggestionChanged(ss []string) {
	c.suggestions, c.suggestionIndex = ss, 0
	if ss == nil {
		c.suggestionIndex = -1
		return
	}
	fmt.Fprintf(c, "[gray::-]%s", ss[c.suggestionIndex])
}

// BufferChanged indicates the buffer was changed.
func (c *Command) BufferChanged(s string) {
	c.update(s)
}

// BufferActive indicates the buff activity changed.
func (c *Command) BufferActive(f bool, k model.BufferKind) {
	if c.activated = f; f {
		c.SetBorder(true)
		c.SetTextColor(c.styles.FgColor())
		c.SetBorderColor(colorFor(k))
		c.icon = iconFor(k)
		c.activate()
	} else {
		c.SetBorder(false)
		c.SetBackgroundColor(c.styles.BgColor())
		c.Clear()
	}
}

func colorFor(k model.BufferKind) tcell.Color {
	switch k {
	case model.Command:
		return tcell.ColorAqua
	default:
		return tcell.ColorSeaGreen
	}
}

func iconFor(k model.BufferKind) rune {
	switch k {
	case model.Command:
		return '🐶'
	default:
		return '🐩'
	}
}
