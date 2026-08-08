// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"sync"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

const maxPromptSuggestions = 8

// SuggestionDropdown renders command suggestions below the prompt.
type SuggestionDropdown struct {
	*tview.Box

	prompt      *Prompt
	styles      *config.Styles
	text        string
	suggestions []string
	selected    int
	active      bool
	mx          sync.RWMutex
}

// NewSuggestionDropdown returns a dropdown view for prompt suggestions.
func NewSuggestionDropdown(p *Prompt, styles *config.Styles) *SuggestionDropdown {
	d := &SuggestionDropdown{
		Box:    tview.NewBox(),
		prompt: p,
		styles: styles,
	}
	d.SetBorder(true)
	d.SetBackgroundColor(styles.BgColor())

	return d
}

// Focus keeps keyboard input on the prompt while the overlay is visible.
func (d *SuggestionDropdown) Focus(delegate func(tview.Primitive)) {
	if delegate == nil || d.prompt == nil {
		return
	}
	if d.prompt.InCmdMode() {
		delegate(d.prompt)
		return
	}
	if d.prompt.app != nil {
		if main := d.prompt.app.Main.GetPrimitive("main"); main != nil {
			delegate(main)
		}
	}
}

// HasFocus returns false since the overlay never owns keyboard input.
func (*SuggestionDropdown) HasFocus() bool {
	return false
}

// Update refreshes the dropdown state.
func (d *SuggestionDropdown) Update(text string, suggestions []string, selected int) {
	d.mx.Lock()
	defer d.mx.Unlock()

	if len(suggestions) == 0 {
		d.active = false
		d.suggestions = nil
		d.selected = -1
		d.text = ""
		return
	}

	d.text = text
	d.suggestions = make([]string, len(suggestions))
	copy(d.suggestions, suggestions)
	d.selected = selected
	if d.selected < 0 || d.selected >= len(d.suggestions) {
		d.selected = 0
	}
	d.active = true
}

// Clear hides the dropdown.
func (d *SuggestionDropdown) Clear() {
	d.mx.Lock()
	defer d.mx.Unlock()

	d.active = false
	d.suggestions = nil
	d.selected = -1
	d.text = ""
}

// IsActive returns true when the dropdown has suggestions to render.
func (d *SuggestionDropdown) IsActive() bool {
	d.mx.RLock()
	defer d.mx.RUnlock()

	return d.active
}

// Items returns the rendered suggestion values.
func (d *SuggestionDropdown) Items() []string {
	d.mx.RLock()
	defer d.mx.RUnlock()

	if len(d.suggestions) == 0 {
		return nil
	}

	items := make([]string, len(d.suggestions))
	for i, s := range d.suggestions {
		items[i] = d.text + s
	}
	return items
}

// SelectedIndex returns the active suggestion index.
func (d *SuggestionDropdown) SelectedIndex() int {
	d.mx.RLock()
	defer d.mx.RUnlock()

	return d.selected
}

// Draw renders the dropdown below the prompt.
func (d *SuggestionDropdown) Draw(screen tcell.Screen) {
	text, suggestions, selected, ok := d.snapshot()
	if !ok || d.prompt == nil {
		return
	}

	px, py, pw, ph := d.prompt.GetRect()
	sw, sh := screen.Size()
	x, y := px+d.prompt.spacer, py+ph-1
	if x >= sw || y >= sh {
		return
	}

	start, rows := suggestionWindow(len(suggestions), selected)
	if y+rows+2 > sh {
		rows = sh - y - 2
	}
	if rows <= 0 {
		return
	}

	width := d.suggestionWidth(text, suggestions[start:start+rows]) + 4
	if maxWidth := pw - d.prompt.spacer - 1; maxWidth > 0 && width > maxWidth {
		width = maxWidth
	}
	if x+width > sw {
		width = sw - x
	}
	if width <= 2 {
		return
	}

	d.SetRect(x, y, width, rows+2)
	d.SetBorderColor(d.styles.Prompt().Border.CommandColor.Color())
	d.SetBorderFocusColor(d.styles.Prompt().Border.CommandColor.Color())
	d.SetBackgroundColor(d.styles.BgColor())
	d.Box.DrawForSubclass(screen, d)

	ix, iy, iw, _ := d.GetInnerRect()
	for row := range rows {
		idx := start + row
		d.drawRow(screen, ix, iy+row, iw, text+suggestions[idx], idx == selected)
	}
}

func (d *SuggestionDropdown) snapshot() (string, []string, int, bool) {
	d.mx.RLock()
	defer d.mx.RUnlock()

	if !d.active || len(d.suggestions) == 0 {
		return "", nil, -1, false
	}

	ss := make([]string, len(d.suggestions))
	copy(ss, d.suggestions)
	return d.text, ss, d.selected, true
}

func (d *SuggestionDropdown) suggestionWidth(text string, suggestions []string) int {
	width := 0
	for _, s := range suggestions {
		if w := tview.TaggedStringWidth(text + s); w > width {
			width = w
		}
	}
	return width
}

func (d *SuggestionDropdown) drawRow(screen tcell.Screen, x, y, width int, text string, selected bool) {
	fg, bg := d.styles.Prompt().FgColor.Color(), d.styles.BgColor()
	if selected {
		fg = d.styles.Table().CursorFgColor.Color()
		bg = d.styles.Table().CursorBgColor.Color()
	}

	style := tcell.StyleDefault.Foreground(fg).Background(bg)
	for col := range width {
		screen.SetContent(x+col, y, ' ', nil, style)
	}
	if width > 2 {
		x++
		width -= 2
	}
	tview.Print(screen, text, x, y, width, tview.AlignLeft, fg)
}

func suggestionWindow(count, selected int) (int, int) {
	if count <= maxPromptSuggestions {
		return 0, count
	}
	if selected < maxPromptSuggestions {
		return 0, maxPromptSuggestions
	}
	return selected - maxPromptSuggestions + 1, maxPromptSuggestions
}
