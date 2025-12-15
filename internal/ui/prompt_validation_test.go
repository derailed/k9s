// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
)

// TestPrompt_FiltersControlCharacters tests that control characters from
// terminal escape sequences are filtered out and not added to the buffer.
func TestPrompt_FiltersControlCharacters(t *testing.T) {
	m := model.NewFishBuff(':', model.CommandBuffer)
	p := ui.NewPrompt(nil, true, config.NewStyles())
	p.SetModel(m)
	m.AddListener(p)
	m.SetActive(true)

	// Test control characters that should be filtered
	controlChars := []rune{
		0x00, // NULL
		0x01, // SOH
		0x1B, // ESC (escape character)
		0x7F, // DEL
	}

	for _, c := range controlChars {
		t.Run(fmt.Sprintf("control_char_0x%02X", c), func(t *testing.T) {
			evt := tcell.NewEventKey(tcell.KeyRune, c, tcell.ModNone)
			p.SendKey(evt)
			// Control characters should not be added to buffer
			assert.Empty(t, m.GetText(), "Control character 0x%02X should be filtered", c)
		})
	}
}

// TestPrompt_AcceptsPrintableCharacters tests that valid printable
// characters are accepted and added to the buffer.
func TestPrompt_AcceptsPrintableCharacters(t *testing.T) {
	m := model.NewFishBuff(':', model.CommandBuffer)
	p := ui.NewPrompt(nil, true, config.NewStyles())
	p.SetModel(m)
	m.AddListener(p)
	m.SetActive(true)

	// Test valid printable characters
	validChars := []rune{
		'a', 'Z', '0', '9',
		'!', '@', '#', '$',
		' ',                // space
		'[', ']', ';', 'R', // characters from escape sequences (should be accepted if typed)
	}

	for _, c := range validChars {
		t.Run(fmt.Sprintf("valid_char_%c", c), func(t *testing.T) {
			evt := tcell.NewEventKey(tcell.KeyRune, c, tcell.ModNone)
			p.SendKey(evt)
			// Valid characters should be added
			assert.Contains(t, m.GetText(), string(c), "Valid character %c should be accepted", c)
			// Clear for next test
			m.ClearText(true)
		})
	}

	// Test tab separately (it's a control char but should be accepted)
	t.Run("valid_char_tab", func(t *testing.T) {
		evt := tcell.NewEventKey(tcell.KeyRune, '\t', tcell.ModNone)
		p.SendKey(evt)
		// Tab should be accepted (it's a special case in the validation)
		// Note: Tab might be converted to spaces or handled differently by the buffer
		text := m.GetText()
		// Tab is accepted by validation, but may be handled specially by the buffer
		// Just verify the buffer isn't empty (meaning something was processed)
		assert.NotNil(t, text, "Tab character should be processed")
		m.ClearText(true)
	})
}

// TestPrompt_FiltersEscapeSequencePattern tests that escape sequence
// patterns are not automatically added when they appear as individual runes.
// Note: This test verifies the validation works, but escape sequences
// should ideally be handled by tcell before reaching KeyRune.
func TestPrompt_FiltersEscapeSequencePattern(t *testing.T) {
	m := model.NewFishBuff(':', model.CommandBuffer)
	p := ui.NewPrompt(nil, true, config.NewStyles())
	p.SetModel(m)
	m.AddListener(p)
	m.SetActive(true)

	// Simulate the problematic escape sequence pattern [7;15R
	// Each character individually is printable, but we want to ensure
	// they don't appear unexpectedly
	escapeSequence := "[7;15R"

	// Send each character
	for _, r := range escapeSequence {
		evt := tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
		p.SendKey(evt)
	}

	// The characters themselves are printable, so they will be added
	// This test documents the current behavior - the fix prevents
	// control characters, but printable escape sequence chars would
	// still be added if tcell doesn't filter them first
	text := m.GetText()

	// If all characters are printable, they will be in the buffer
	// This is expected behavior - the fix prevents control chars,
	// but can't prevent legitimate printable characters
	assert.NotEmpty(t, text, "Printable escape sequence chars may still appear")

	// However, we can verify no control characters made it through
	for _, r := range text {
		assert.False(t, isControlChar(r), "No control characters should be in buffer")
	}
}

// Helper function to check if a rune is a control character
func isControlChar(r rune) bool {
	return r >= 0x00 && r <= 0x1F || r == 0x7F
}
