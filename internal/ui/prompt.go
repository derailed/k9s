package ui

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell/v2"
)

const (
	defaultPrompt = "%c> [::b]%s"
	defaultSpacer = 4
)

var (
	_ PromptModel = (*model.FishBuff)(nil)
	_ Suggester   = (*model.FishBuff)(nil)
)

// Suggester provides suggestions.
type Suggester interface {
	// CurrentSuggestion returns the current suggestion.
	CurrentSuggestion() (string, bool)

	// NextSuggestion returns the next suggestion.
	NextSuggestion() (string, bool)

	// PrevSuggestion returns the prev suggestion.
	PrevSuggestion() (string, bool)

	// ClearSuggestions clear out all suggestions.
	ClearSuggestions()
}

// PromptModel represents a prompt buffer
type PromptModel interface {
	// SetText sets the model text.
	SetText(string)

	// GetText returns the current text.
	GetText() string

	// ClearText clears out model text.
	ClearText(fire bool)

	// Notify notifies all listener of current suggestions.
	Notify(bool)

	// AddListener registers a command listener.
	AddListener(model.BuffWatcher)

	// RemoveListener removes a listener.
	RemoveListener(model.BuffWatcher)

	// IsActive returns true if prompt is active.
	IsActive() bool

	// SetActive sets whether the prompt is active or not.
	SetActive(bool)

	// Add adds a new char to the prompt.
	Add(rune)

	// Delete deletes the last prompt character.
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
	p := Prompt{
		styles:   styles,
		noIcons:  noIcons,
		TextView: tview.NewTextView(),
		spacer:   defaultSpacer,
	}
	if noIcons {
		p.spacer--
	}
	p.SetWordWrap(true)
	p.SetWrap(true)
	p.SetDynamicColors(true)
	p.SetBorder(true)
	p.SetBorderPadding(0, 0, 1, 1)
	p.SetBackgroundColor(styles.BgColor())
	p.SetTextColor(styles.FgColor())
	styles.AddListener(&p)
	p.SetInputCapture(p.keyboard)

	return &p
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

// SetModel sets the prompt buffer model.
func (p *Prompt) SetModel(m PromptModel) {
	if p.model != nil {
		p.model.RemoveListener(p)
	}
	p.model = m
	p.model.AddListener(p)
}

func (p *Prompt) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	m, ok := p.model.(Suggester)
	if !ok {
		return evt
	}

	switch evt.Key() {
	case tcell.KeyBackspace2, tcell.KeyBackspace, tcell.KeyDelete:
		p.model.Delete()
	case tcell.KeyRune:
		p.model.Add(evt.Rune())
	case tcell.KeyEscape:
		p.model.ClearText(true)
		p.model.SetActive(false)
	case tcell.KeyEnter, tcell.KeyCtrlE:
		p.model.SetText(p.model.GetText())
		p.model.SetActive(false)
	case tcell.KeyCtrlW, tcell.KeyCtrlU:
		p.model.ClearText(true)
	case tcell.KeyUp:
		if s, ok := m.NextSuggestion(); ok {
			p.suggest(p.model.GetText(), s)
		}
	case tcell.KeyDown:
		if s, ok := m.PrevSuggestion(); ok {
			p.suggest(p.model.GetText(), s)
		}
	case tcell.KeyTab, tcell.KeyRight, tcell.KeyCtrlF:
		if s, ok := m.CurrentSuggestion(); ok {
			p.model.SetText(p.model.GetText() + s)
			m.ClearSuggestions()
		}
	}

	return nil
}

// StylesChanged notifies skin changed.
func (p *Prompt) StylesChanged(s *config.Styles) {
	p.styles = s
	p.SetBackgroundColor(s.BgColor())
	p.SetTextColor(s.FgColor())
}

// InCmdMode returns true if command is active, false otherwise.
func (p *Prompt) InCmdMode() bool {
	if p.model == nil {
		return false
	}
	return p.model.IsActive()
}

func (p *Prompt) activate() {
	p.SetCursorIndex(len(p.model.GetText()))
	p.write(p.model.GetText(), "")
	p.model.Notify(false)
}

func (p *Prompt) update(s string) {
	p.Clear()
	p.write(s, "")
}

func (p *Prompt) suggest(text, suggestion string) {
	p.Clear()
	p.write(text, suggestion)
}

func (p *Prompt) write(text, suggest string) {
	p.SetCursorIndex(p.spacer + len(text))
	txt := text
	if suggest != "" {
		txt += "[gray::-]" + suggest
	}
	fmt.Fprintf(p, defaultPrompt, p.icon, txt)
}

// ----------------------------------------------------------------------------
// Event Listener protocol...

// BufferCompleted indicates input was accepted.
func (p *Prompt) BufferCompleted(s string) {
	p.update(s)
}

// BufferChanged indicates the buffer was changed.
func (p *Prompt) BufferChanged(s string) {
	p.update(s)
}

// SuggestionChanged notifies the suggestion changed.
func (p *Prompt) SuggestionChanged(text, sugg string) {
	p.Clear()
	p.write(text, sugg)
}

// BufferActive indicates the buff activity changed.
func (p *Prompt) BufferActive(activate bool, kind model.BufferKind) {
	if activate {
		p.ShowCursor(true)
		p.SetBorder(true)
		p.SetTextColor(p.styles.FgColor())
		p.SetBorderColor(colorFor(kind))
		p.icon = p.iconFor(kind)
		p.activate()
		return
	}

	p.ShowCursor(false)
	p.SetBorder(false)
	p.SetBackgroundColor(p.styles.BgColor())
	p.Clear()
}

func (p *Prompt) iconFor(k model.BufferKind) rune {
	if p.noIcons {
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
