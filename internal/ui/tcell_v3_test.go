// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell/v3"
	"github.com/stretchr/testify/assert"
)

func TestAsKeyTcellV3(t *testing.T) {
	tests := []struct {
		name  string
		event *tcell.EventKey
		want  tcell.Key
	}{
		{"lowercase", tcell.NewEventKey(tcell.KeyRune, "a", tcell.ModNone), ui.KeyA},
		{"uppercase", tcell.NewEventKey(tcell.KeyRune, "A", tcell.ModNone), ui.KeyShiftA},
		{"legacy ctrl-a", tcell.NewEventKey(tcell.KeyCtrlA, "", tcell.ModCtrl), tcell.KeyCtrlA},
		{"advanced ctrl-a", tcell.NewEventKeyEx(tcell.KeyRune, "a", tcell.ModCtrl, true, 0, 1), tcell.KeyCtrlA},
		{"advanced ctrl-space", tcell.NewEventKeyEx(tcell.KeyRune, " ", tcell.ModCtrl, true, 0, 1), ui.KeyCtrlSpace},
		{"advanced ctrl-backslash", tcell.NewEventKeyEx(tcell.KeyRune, "\\", tcell.ModCtrl, true, 0, 1), ui.KeyCtrlBackslash},
		{"shift tab", tcell.NewEventKey(tcell.KeyTab, "", tcell.ModShift), tcell.KeyBacktab},
		{"legacy backtab", tcell.NewEventKey(tcell.KeyBacktab, "", tcell.ModNone), tcell.KeyBacktab},
		{"grapheme is not shortcut", tcell.NewEventKey(tcell.KeyRune, "a\u0301", tcell.ModNone), tcell.KeyRune},
		{"empty key", tcell.NewEventKey(tcell.KeyRune, "", tcell.ModNone), tcell.KeyRune},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, ui.AsKey(tt.event))
		})
	}
	assert.NotEqual(t, ui.KeyShiftA, tcell.KeyCtrlA)
	assert.Equal(t, "Shift-A", tcell.KeyNames[ui.KeyShiftA])
	assert.Equal(t, "Ctrl-A", tcell.KeyNames[tcell.KeyCtrlA])
}

func TestPromptTcellV3TextAndControlKeys(t *testing.T) {
	m := model.NewFishBuff(':', model.CommandBuffer)
	p := ui.NewPrompt(nil, true, config.NewStyles())
	p.SetModel(m)
	m.SetActive(true)

	p.SendKey(tcell.NewEventKey(tcell.KeyRune, "a\u0301", tcell.ModNone))
	assert.Equal(t, "a\u0301", m.GetText())
	p.SendKey(tcell.NewEventKeyEx(tcell.KeyRune, " ", tcell.ModCtrl, true, 0, 1))
	assert.Equal(t, "a\u0301", m.GetText(), "Ctrl-Space must not insert a space")
	p.SendKey(tcell.NewEventKeyEx(tcell.KeyRune, "u", tcell.ModCtrl, true, 0, 1))
	assert.Empty(t, m.GetText(), "advanced Ctrl-U must clear the prompt")
	p.SendKey(tcell.NewEventKey(tcell.KeyRune, "x", tcell.ModNone))
	p.SendKey(tcell.NewEventKey(tcell.KeyBackspace, "", tcell.ModNone))
	assert.Empty(t, m.GetText())
}
