// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"os"
	"testing"
	"time"

	"github.com/derailed/tview"
	"github.com/gdamore/tcell/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Run in a real PTY with K9S_TERMINAL_TEST=1 and TERM=xterm-direct (or
// another *-direct terminal). Ordinary unit tests must not acquire the TTY.
func TestDirectTerminal(t *testing.T) {
	if os.Getenv("K9S_TERMINAL_TEST") != "1" {
		t.Skip("requires an explicitly requested real-PTY terminal test")
	}
	type frame struct {
		colors                 int
		text                   string
		foreground, background tcell.Color
	}
	frames := make(chan frame, 1)
	app := tview.NewApplication()
	view := tview.NewTextView().SetText("K9S-DIRECT-OK")
	view.SetTextColor(tcell.NewHexColor(0xabcdef))
	view.SetBackgroundColor(tcell.NewHexColor(0x123456))
	app.SetRoot(view, true).SetAfterDrawFunc(func(screen tcell.Screen) {
		text, style, _ := screen.Get(0, 0)
		fg, bg := style.GetForeground(), style.GetBackground()
		select {
		case frames <- frame{screen.Colors(), text, fg, bg}:
		default:
		}
	})
	done := make(chan error, 1)
	go func() { done <- app.Run() }()
	select {
	case got := <-frames:
		app.Stop()
		require.NoError(t, <-done)
		assert.Greater(t, got.colors, 256, "*-direct must support RGB colors")
		assert.Equal(t, "K", got.text)
		assert.Equal(t, tcell.NewHexColor(0xabcdef), got.foreground)
		assert.Equal(t, tcell.NewHexColor(0x123456), got.background)
	case err := <-done:
		t.Fatalf("terminal application exited before drawing: %v", err)
	case <-time.After(10 * time.Second):
		t.Fatal("terminal application did not render within 10 seconds")
	}
}
