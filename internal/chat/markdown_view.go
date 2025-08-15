// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package chat

import (
	"io"
	"strings"

	"github.com/charmbracelet/glamour"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// MarkdownTextView is a custom tview component that can render markdown with ANSI colors.
type MarkdownTextView struct {
	*tview.Box

	text     strings.Builder
	renderer *glamour.TermRenderer
	lines    []string

	// Scrolling
	lineOffset int

	// Styling
	textColor tcell.Color
	bgColor   tcell.Color
}

// NewMarkdownTextView creates a new markdown-capable text view.
func NewMarkdownTextView() *MarkdownTextView {
	mv := &MarkdownTextView{
		Box:       tview.NewBox(),
		textColor: tcell.ColorWhite,
		bgColor:   tcell.ColorDefault,
		lines:     make([]string, 0),
	}

	// Initialize glamour renderer
	var err error
	mv.renderer, err = glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(60),
	)
	if err != nil {
		// Fallback renderer
		mv.renderer, _ = glamour.NewTermRenderer()
	}

	return mv
}

// SetText sets the text content, processing it as markdown.
func (mv *MarkdownTextView) SetText(text string) *MarkdownTextView {
	mv.text.Reset()
	mv.text.WriteString(text)
	mv.updateLines()
	return mv
}

// GetText returns the current text content.
func (mv *MarkdownTextView) GetText(stripTags bool) string {
	if stripTags {
		// Return plain text without ANSI codes
		return mv.stripANSI(mv.text.String())
	}
	return mv.text.String()
}

// AddText appends text to the current content.
func (mv *MarkdownTextView) AddText(text string) *MarkdownTextView {
	mv.text.WriteString(text)
	mv.updateLines()
	return mv
}

// AddMarkdown adds markdown content, rendering it with Glamour.
func (mv *MarkdownTextView) AddMarkdown(markdown string) *MarkdownTextView {
	rendered, err := mv.renderer.Render(markdown)
	if err != nil {
		// Fallback to plain text
		rendered = markdown
	}
	mv.text.WriteString(rendered)
	mv.updateLines()
	return mv
}

// ScrollToEnd scrolls to the bottom of the content.
func (mv *MarkdownTextView) ScrollToEnd() *MarkdownTextView {
	_, _, _, height := mv.GetInnerRect()
	if len(mv.lines) > height {
		mv.lineOffset = len(mv.lines) - height
	}
	return mv
}

// Clear clears all content.
func (mv *MarkdownTextView) Clear() *MarkdownTextView {
	mv.text.Reset()
	mv.lines = mv.lines[:0]
	mv.lineOffset = 0
	return mv
}

// updateLines processes the text content and splits it into renderable lines.
func (mv *MarkdownTextView) updateLines() {
	content := mv.text.String()
	mv.lines = strings.Split(content, "\n")
}

// stripANSI removes ANSI escape sequences from text.
func (mv *MarkdownTextView) stripANSI(text string) string {
	// Simple ANSI stripping - could be more sophisticated
	lines := strings.Split(text, "\n")
	var cleaned []string

	for _, line := range lines {
		// Remove common ANSI escape sequences
		clean := line
		// Remove color codes like \x1b[38;5;252m
		for strings.Contains(clean, "\x1b[") {
			start := strings.Index(clean, "\x1b[")
			if start == -1 {
				break
			}
			end := strings.Index(clean[start:], "m")
			if end == -1 {
				break
			}
			clean = clean[:start] + clean[start+end+1:]
		}
		cleaned = append(cleaned, clean)
	}

	return strings.Join(cleaned, "\n")
}

// Draw renders the component.
func (mv *MarkdownTextView) Draw(screen tcell.Screen) {
	mv.Box.DrawForSubclass(screen, mv)

	x, y, width, height := mv.GetInnerRect()

	// Render visible lines
	for i := 0; i < height && mv.lineOffset+i < len(mv.lines); i++ {
		lineIndex := mv.lineOffset + i
		if lineIndex >= len(mv.lines) {
			break
		}

		line := mv.lines[lineIndex]
		mv.drawLine(screen, x, y+i, width, line)
	}
}

// drawLine renders a single line with ANSI color support.
func (mv *MarkdownTextView) drawLine(screen tcell.Screen, x, y, width int, text string) {
	// This is where the magic happens - we parse ANSI codes and render them
	currentX := x
	currentColor := mv.textColor
	currentBg := mv.bgColor
	bold := false

	i := 0
	runes := []rune(text)

	for i < len(runes) && currentX < x+width {
		r := runes[i]

		// Check for ANSI escape sequence
		if r == '\x1b' && i+1 < len(runes) && runes[i+1] == '[' {
			// Parse ANSI sequence
			start := i
			i += 2 // Skip \x1b[

			// Find the end of the sequence (letter)
			for i < len(runes) && (runes[i] >= '0' && runes[i] <= '9' || runes[i] == ';') {
				i++
			}

			if i < len(runes) {
				cmd := runes[i]
				sequence := string(runes[start : i+1])

				// Parse color codes
				if cmd == 'm' {
					currentColor, currentBg, bold = mv.parseColorCode(sequence, currentColor, currentBg, bold)
				}
				i++
			}
		} else {
			// Regular character - render it
			style := tcell.StyleDefault.Foreground(currentColor).Background(currentBg)
			if bold {
				style = style.Bold(true)
			}

			screen.SetContent(currentX, y, r, nil, style)
			currentX++
			i++
		}
	}
}

// parseColorCode parses ANSI color codes and returns tcell colors.
func (mv *MarkdownTextView) parseColorCode(sequence string, currentFg, currentBg tcell.Color, bold bool) (tcell.Color, tcell.Color, bool) {
	// Simple color parsing - could be expanded
	switch {
	case strings.Contains(sequence, "0m"):
		// Reset
		return mv.textColor, mv.bgColor, false
	case strings.Contains(sequence, "1m"):
		// Bold
		return currentFg, currentBg, true
	case strings.Contains(sequence, "31m"):
		// Red
		return tcell.ColorRed, currentBg, bold
	case strings.Contains(sequence, "32m"):
		// Green
		return tcell.ColorGreen, currentBg, bold
	case strings.Contains(sequence, "33m"):
		// Yellow
		return tcell.ColorYellow, currentBg, bold
	case strings.Contains(sequence, "34m"):
		// Blue
		return tcell.ColorBlue, currentBg, bold
	case strings.Contains(sequence, "35m"):
		// Magenta
		return tcell.ColorPurple, currentBg, bold
	case strings.Contains(sequence, "36m"):
		// Cyan
		return tcell.ColorTeal, currentBg, bold
	case strings.Contains(sequence, "37m"):
		// White
		return tcell.ColorWhite, currentBg, bold
	case strings.Contains(sequence, "38;5;"):
		// 256-color mode (simplified)
		if strings.Contains(sequence, "252m") {
			return tcell.ColorLightGray, currentBg, bold
		}
		return currentFg, currentBg, bold
	default:
		return currentFg, currentBg, bold
	}
}

// InputHandler handles input events.
func (mv *MarkdownTextView) InputHandler() func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
	return mv.WrapInputHandler(func(event *tcell.EventKey, setFocus func(p tview.Primitive)) {
		switch event.Key() {
		case tcell.KeyUp:
			if mv.lineOffset > 0 {
				mv.lineOffset--
			}
		case tcell.KeyDown:
			_, _, _, height := mv.GetInnerRect()
			if mv.lineOffset < len(mv.lines)-height {
				mv.lineOffset++
			}
		case tcell.KeyPgUp:
			_, _, _, height := mv.GetInnerRect()
			mv.lineOffset -= height
			if mv.lineOffset < 0 {
				mv.lineOffset = 0
			}
		case tcell.KeyPgDn:
			_, _, _, height := mv.GetInnerRect()
			mv.lineOffset += height
			if mv.lineOffset > len(mv.lines)-height {
				mv.lineOffset = len(mv.lines) - height
			}
			if mv.lineOffset < 0 {
				mv.lineOffset = 0
			}
		}
	})
}

// MouseHandler handles mouse events.
func (mv *MarkdownTextView) MouseHandler() func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
	return mv.WrapMouseHandler(func(action tview.MouseAction, event *tcell.EventMouse, setFocus func(p tview.Primitive)) (consumed bool, capture tview.Primitive) {
		if action == tview.MouseLeftClick {
			setFocus(mv)
			return true, nil
		}
		return false, nil
	})
}

// Write implements io.Writer for compatibility.
func (mv *MarkdownTextView) Write(p []byte) (n int, err error) {
	mv.AddText(string(p))
	return len(p), nil
}

// Ensure MarkdownTextView implements io.Writer
var _ io.Writer = (*MarkdownTextView)(nil)
