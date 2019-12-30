package ui

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

const defaultPrompt = "%c> %s"

// Command captures users free from command input.
type Command struct {
	*tview.TextView

	activated bool
	icon      rune
	text      string
	styles    *config.Styles
}

// NewCommand returns a new command view.
func NewCommand(styles *config.Styles) *Command {
	c := Command{styles: styles, TextView: tview.NewTextView()}
	c.SetWordWrap(true)
	c.SetWrap(true)
	c.SetDynamicColors(true)
	c.SetBorder(true)
	c.SetBorderPadding(0, 0, 1, 1)
	c.SetBackgroundColor(styles.BgColor())
	c.SetTextColor(styles.FgColor())
	styles.AddListener(&c)

	return &c
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

func (c *Command) write(s string) {
	fmt.Fprintf(c, defaultPrompt, c.icon, s)
}

// ----------------------------------------------------------------------------
// Event Listener protocol...

// BufferChanged indicates the buffer was changed.
func (c *Command) BufferChanged(s string) {
	c.update(s)
}

// BufferActive indicates the buff activity changed.
func (c *Command) BufferActive(f bool, k BufferKind) {
	if c.activated = f; f {
		c.SetBorder(true)
		c.SetTextColor(c.styles.FgColor())
		c.SetBorderColor(colorFor(k))
		c.icon = iconFor(k)
		// c.reset()
		c.activate()
	} else {
		c.SetBorder(false)
		c.SetBackgroundColor(c.styles.BgColor())
		c.Clear()
	}
}

func colorFor(k BufferKind) tcell.Color {
	switch k {
	case CommandBuff:
		return tcell.ColorAqua
	default:
		return tcell.ColorSeaGreen
	}
}

func iconFor(k BufferKind) rune {
	switch k {
	case CommandBuff:
		return '🐶'
	default:
		return '🐩'
	}
}
