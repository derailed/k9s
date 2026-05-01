// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"strings"
	"testing"

	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimScreen_BasicText(t *testing.T) {
	sim := NewSimScreen(t, 40, 5)
	defer sim.Fini()

	tv := tview.NewTextView()
	tv.SetText("hello world")
	sim.Render(tv)

	text := sim.ScreenText()
	assert.Contains(t, text, "hello world")
}

func TestSimScreen_BracketLiteralRendering(t *testing.T) {
	sim := NewSimScreen(t, 80, 10)
	defer sim.Fini()

	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	// tview escaping: [[ produces a literal [ on screen
	tv.SetText("[[INFO]] server started on [[8080]]")
	sim.Render(tv)

	lines := sim.ScreenLines()
	require.Positive(t, len(lines))
	assert.Contains(t, lines[0], "[INFO]", "literal brackets should render on screen")
	assert.Contains(t, lines[0], "[8080]", "literal brackets should render on screen")
}

func TestSimScreen_MultipleLinesRendering(t *testing.T) {
	sim := NewSimScreen(t, 60, 10)
	defer sim.Fini()

	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetText("line one\nline two\nline three")
	sim.Render(tv)

	lines := sim.ScreenLines()
	found := 0
	for _, l := range lines {
		trimmed := strings.TrimSpace(l)
		switch trimmed {
		case "line one", "line two", "line three":
			found++
		}
	}
	assert.Equal(t, 3, found, "all three lines should be visible")
}

func TestSimScreen_PlainTextNoBrackets(t *testing.T) {
	sim := NewSimScreen(t, 80, 5)
	defer sim.Fini()

	tv := tview.NewTextView()
	tv.SetText("error message without any special markup")
	sim.Render(tv)

	text := sim.ScreenText()
	assert.Contains(t, text, "error message without any special markup")
}

func TestSimScreen_EscapedBracketsInLogLine(t *testing.T) {
	sim := NewSimScreen(t, 100, 5)
	defer sim.Fini()

	// Simulate what sanitizeEsc produces: [INFO[] becomes [INFO] on screen
	// But tview renders [INFO[] by stripping the first bracket pair as a tag.
	// The correct way to show literal [INFO] is [[INFO]]
	tv := tview.NewTextView()
	tv.SetDynamicColors(true)
	tv.SetText("2024-01-01 [[INFO]] Application started")
	sim.Render(tv)

	lines := sim.ScreenLines()
	require.Positive(t, len(lines))
	assert.Contains(t, lines[0], "[INFO]")
	assert.Contains(t, lines[0], "Application started")
}

func TestSimScreen_ScreenLines(t *testing.T) {
	sim := NewSimScreen(t, 20, 3)
	defer sim.Fini()

	tv := tview.NewTextView()
	tv.SetText("abc\nxyz")
	sim.Render(tv)

	lines := sim.ScreenLines()
	assert.Equal(t, "abc", lines[0])
	assert.Equal(t, "xyz", lines[1])
}
