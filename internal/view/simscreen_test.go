// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"strings"
	"testing"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// SimScreen wraps a tcell.SimulationScreen and a tview.Application for
// rendering TUI components in tests without a real terminal.
//
// Usage:
//
//	sim := NewSimScreen(t, 80, 25)
//	defer sim.Fini()
//	tv := tview.NewTextView()
//	tv.SetText("hello")
//	sim.Render(tv)
//	text := sim.ScreenText()
//	assert.Contains(t, text, "hello")
type SimScreen struct {
	t      testing.TB
	Screen tcell.SimulationScreen
	App    *tview.Application
	width  int
	height int
}

// NewSimScreen creates a SimulationScreen of the given dimensions
// and attaches it to a new tview.Application. The screen is initialized
// and ready for rendering.
func NewSimScreen(t testing.TB, width, height int) *SimScreen {
	t.Helper()
	scr := tcell.NewSimulationScreen("UTF-8")
	if err := scr.Init(); err != nil {
		t.Fatalf("SimulationScreen.Init: %v", err)
	}
	scr.SetSize(width, height)

	app := tview.NewApplication()
	app.SetScreen(scr)

	return &SimScreen{
		t:      t,
		Screen: scr,
		App:    app,
		width:  width,
		height: height,
	}
}

// Render draws the given Primitive into the simulation screen.
// It sets the primitive as root, forces a full-screen layout, and
// calls Draw() to flush the cell buffer.
func (s *SimScreen) Render(p tview.Primitive) {
	s.t.Helper()
	s.App.SetRoot(p, true)
	p.SetRect(0, 0, s.width, s.height)
	s.App.ForceDraw()
}

// ScreenText reads every cell from the simulation screen and returns
// the visible text as a single string with embedded newlines.
// Trailing whitespace on each row is trimmed.
func (s *SimScreen) ScreenText() string {
	s.t.Helper()
	cells, w, h := s.Screen.GetContents()
	var sb strings.Builder
	for row := 0; row < h; row++ {
		for col := 0; col < w; col++ {
			cell := cells[row*w+col]
			if len(cell.Runes) > 0 {
				sb.WriteRune(cell.Runes[0])
			} else {
				sb.WriteByte(' ')
			}
		}
		sb.WriteByte('\n')
	}
	return sb.String()
}

// ScreenLines returns each row of the screen as a trimmed string.
func (s *SimScreen) ScreenLines() []string {
	raw := s.ScreenText()
	rows := strings.Split(raw, "\n")
	out := make([]string, 0, len(rows))
	for _, r := range rows {
		out = append(out, strings.TrimRight(r, " "))
	}
	return out
}

// CellStyle returns the style applied to the cell at (col, row).
func (s *SimScreen) CellStyle(col, row int) tcell.Style {
	s.t.Helper()
	cells, w, _ := s.Screen.GetContents()
	return cells[row*w+col].Style
}

// Fini tears down the simulation screen.
func (s *SimScreen) Fini() {
	s.Screen.Fini()
}
