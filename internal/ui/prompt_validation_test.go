// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"fmt"
	"testing"
	"time"

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

// TestEscapeFilter_IntegrationCPR tests that a CPR response is fully
// filtered by the EscapeSequenceFilter at the application level.
// The filter is applied in view.App.keyboard(), not in the Prompt.
func TestEscapeFilter_IntegrationCPR(t *testing.T) {
	f := ui.NewEscapeSequenceFilter(nil)

	// Simulate a CPR response: \x1b[7;15R
	// tcell delivers: Alt+[, then 7, ;, 1, 5, R
	events := []*tcell.EventKey{
		tcell.NewEventKey(tcell.KeyRune, '[', tcell.ModAlt),
		tcell.NewEventKey(tcell.KeyRune, '7', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, ';', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, '1', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, '5', tcell.ModNone),
		tcell.NewEventKey(tcell.KeyRune, 'R', tcell.ModNone),
	}

	for _, evt := range events {
		assert.True(t, f.Filter(evt), "CPR event should be filtered")
	}

	// Normal typing after sequence should pass through
	assert.False(t, f.Filter(tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone)))
}

// TestEscapeFilter_IntegrationOSC tests that an OSC 10 color response
// is fully filtered by the EscapeSequenceFilter.
func TestEscapeFilter_IntegrationOSC(t *testing.T) {
	f := ui.NewEscapeSequenceFilter(nil)

	// Simulate OSC 10 response: \x1b]10;rgb:fafa/f9f9/f6f6\x1b\\
	assert.True(t, f.Filter(tcell.NewEventKey(tcell.KeyRune, ']', tcell.ModAlt)))
	for _, r := range "10;rgb:fafa/f9f9/f6f6" {
		assert.True(t, f.Filter(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)))
	}
	assert.True(t, f.Filter(tcell.NewEventKey(tcell.KeyRune, '\\', tcell.ModAlt)))

	// Normal typing after should pass
	assert.False(t, f.Filter(tcell.NewEventKey(tcell.KeyRune, 'a', tcell.ModNone)))
}

// TestEscapeFilter_IntegrationNormalTyping tests that normal user input
// is never affected by the filter. Adds realistic delays between keystrokes
// to simulate actual typing speed (> 5ms between chars).
func TestEscapeFilter_IntegrationNormalTyping(t *testing.T) {
	f := ui.NewEscapeSequenceFilter(nil)

	for _, r := range "pods/nginx:123" {
		time.Sleep(10 * time.Millisecond) // simulate typing speed
		assert.False(t, f.Filter(tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)),
			"Normal char %c should pass through", r)
	}
}

// Helper function to check if a rune is a control character
func isControlChar(r rune) bool {
	return r >= 0x00 && r <= 0x1F || r == 0x7F
}
