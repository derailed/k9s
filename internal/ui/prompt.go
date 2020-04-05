package ui

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

const (
	defaultPrompt = "%c> [::b]%s"
	defaultSpacer = 4
)

var _ PromptModel = (*model.CmdBuff)(nil)
var _ Suggester = (*model.CmdBuff)(nil)
var _ PromptModel = (*model.FishBuff)(nil)
var _ Suggester = (*model.FishBuff)(nil)

type Suggester interface {
	CurrentSuggestion() (string, bool)
	NextSuggestion() (string, bool)
	PrevSuggestion() (string, bool)
	ClearSuggestions()
}

type PromptModel interface {
	// AutoSuggests returns true if model implements auto suggestions.
	AutoSuggests() bool

	// Suggestions returns suggestions.
	Suggestions() []string

	// SetText sets the model text.
	SetText(string)

	// GetText returns the current text.
	GetText() string

	// ClearText clears out model text.
	ClearText()

	// Notify notifies all listener of current suggestions.
	Notify()

	// AddListener registers a command listener.
	AddListener(model.BuffWatcher)

	// RemoveListener removes a listener.
	RemoveListener(model.BuffWatcher)

	IsActive() bool
	SetActive(bool)
	Add(rune)
	Delete()
}

// Prompt captures users free from command input.
type Prompt struct {
	*tview.TextView

	noIcons bool
	icon    rune
	styles  *config.Styles
	model   PromptModel
	spacer  int
}

// NewPrompt returns a new command view.
func NewPrompt(noIcons bool, styles *config.Styles) *Prompt {
	c := Prompt{
		styles:   styles,
		noIcons:  noIcons,
		TextView: tview.NewTextView(),
		spacer:   defaultSpacer,
	}
	if noIcons {
		c.spacer--
	}
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

// SendKey sends an keyboard event (testing only!).
func (p *Prompt) SendKey(evt *tcell.EventKey) {
	p.keyboard(evt)
}

// SendStrokes (testing only!)
func (p *Prompt) SendStrokes(s string) {
	for _, r := range s {
		p.keyboard(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone))
	}
}

func (c *Prompt) SetModel(m PromptModel) {
	if c.model != nil {
		c.model.RemoveListener(c)
	}
	c.model = m
	c.model.AddListener(c)
}

func (c *Prompt) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	m, ok := c.model.(Suggester)
	if !ok {
		return nil
	}

	switch evt.Key() {
	case tcell.KeyBackspace2, tcell.KeyBackspace, tcell.KeyDelete:
		c.model.Delete()
	case tcell.KeyRune:
		c.model.Add(evt.Rune())
	case tcell.KeyEscape:
		c.model.ClearText()
		c.model.SetActive(false)
	case tcell.KeyEnter, tcell.KeyCtrlE:
		if curr, ok := m.CurrentSuggestion(); ok {
			c.model.SetText(c.model.GetText() + curr)
		}
		c.model.SetActive(false)
	case tcell.KeyCtrlW, tcell.KeyCtrlU:
		c.model.ClearText()
	case tcell.KeyDown:
		if next, ok := m.NextSuggestion(); ok {
			c.suggest(c.model.GetText(), next)
		}
	case tcell.KeyUp:
		if prev, ok := m.PrevSuggestion(); ok {
			c.suggest(c.model.GetText(), prev)
		}
	case tcell.KeyTab, tcell.KeyRight, tcell.KeyCtrlF:
		if curr, ok := m.CurrentSuggestion(); ok {
			c.model.SetText(c.model.GetText() + curr)
			m.ClearSuggestions()
		}
	}
	return evt
}

// StylesChanged notifies skin changed.
func (c *Prompt) StylesChanged(s *config.Styles) {
	c.styles = s
	c.SetBackgroundColor(s.BgColor())
	c.SetTextColor(s.FgColor())
}

// InCmdMode returns true if command is active, false otherwise.
func (c *Prompt) InCmdMode() bool {
	if c.model == nil {
		return false
	}
	return c.model.IsActive()
}

func (c *Prompt) activate() {
	c.SetCursorIndex(len(c.model.GetText()))
	c.write(c.model.GetText(), "")
	c.model.Notify()
}

func (c *Prompt) update(s string) {
	c.Clear()
	c.write(s, "")
}

func (c *Prompt) suggest(text, suggestion string) {
	c.Clear()
	c.write(text, suggestion)
}

func (c *Prompt) write(text, suggest string) {
	c.SetCursorIndex(c.spacer + len(text))
	txt := text
	if suggest != "" {
		txt += "[gray::-]" + suggest
	}
	fmt.Fprintf(c, defaultPrompt, c.icon, txt)
}

// ----------------------------------------------------------------------------
// Event Listener protocol...

// BufferChanged indicates the buffer was changed.
func (c *Prompt) BufferChanged(s string) {
	c.update(s)
}

func (c *Prompt) SuggestionChanged(text, sugg string) {
	c.Clear()
	c.write(text, sugg)
}

// BufferActive indicates the buff activity changed.
func (c *Prompt) BufferActive(activate bool, kind model.BufferKind) {
	if activate {
		c.ShowCursor(true)
		c.SetBorder(true)
		c.SetTextColor(c.styles.FgColor())
		c.SetBorderColor(colorFor(kind))
		c.icon = c.iconFor(kind)
		c.activate()
		return
	}

	c.ShowCursor(false)
	c.SetBorder(false)
	c.SetBackgroundColor(c.styles.BgColor())
	c.Clear()
}

func (c *Prompt) iconFor(k model.BufferKind) rune {
	if c.noIcons {
		return ' '
	}

	switch k {
	case model.CommandBuffer:
		return '🐶'
	default:
		return '🐩'
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func colorFor(k model.BufferKind) tcell.Color {
	switch k {
	case model.CommandBuffer:
		return tcell.ColorAqua
	default:
		return tcell.ColorSeaGreen
	}
}
