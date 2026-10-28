// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasteScreenIsPasted(t *testing.T) {
	sim := tcell.NewSimulationScreen("UTF-8")
	require.NoError(t, sim.Init())
	defer sim.Fini()
	s := ui.NewPasteScreen(sim)

	typed := tcell.NewEventKey(tcell.KeyRune, 'u', tcell.ModNone)
	pasted1 := tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone)
	pasted2 := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	after := tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone)
	for _, evt := range []tcell.Event{
		typed,
		tcell.NewEventPaste(true),
		pasted1,
		pasted2,
		tcell.NewEventPaste(false),
		after,
	} {
		require.NoError(t, sim.PostEvent(evt))
		assert.Equal(t, evt, s.PollEvent())
	}

	assert.False(t, s.IsPasted(typed))
	assert.True(t, s.IsPasted(pasted1))
	assert.True(t, s.IsPasted(pasted2))
	assert.False(t, s.IsPasted(after))

	// Each pasted event is only reported once.
	assert.False(t, s.IsPasted(pasted1))
}

func TestPastedKey(t *testing.T) {
	r := tcell.NewEventKey(tcell.KeyRune, 's', tcell.ModNone)
	enter := tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone)
	prompt := ui.NewPrompt(nil, true, config.NewStyles())

	uu := map[string]struct {
		focus tview.Primitive
		evt   *tcell.EventKey
		e     *tcell.EventKey
	}{
		"table-rune":   {focus: tview.NewTable(), evt: r},
		"table-enter":  {focus: tview.NewTable(), evt: enter},
		"no-focus":     {focus: nil, evt: r},
		"input-rune":   {focus: tview.NewInputField(), evt: r, e: r},
		"input-enter":  {focus: tview.NewInputField(), evt: enter},
		"prompt-rune":  {focus: prompt, evt: r, e: r},
		"prompt-enter": {focus: prompt, evt: enter},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, ui.PastedKey(u.focus, u.evt))
		})
	}
}
