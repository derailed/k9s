// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"sync"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// PasteScreen wraps a screen to track which key events are part of a bracketed paste.
// Pasted text arrives as individual key events, so without this it would be
// interpreted as hotkeys.
type PasteScreen struct {
	tcell.Screen

	mx      sync.Mutex
	pasting bool
	pasted  map[*tcell.EventKey]struct{}
}

// NewPasteScreen returns a new paste aware screen.
func NewPasteScreen(s tcell.Screen) *PasteScreen {
	return &PasteScreen{
		Screen: s,
		pasted: make(map[*tcell.EventKey]struct{}),
	}
}

// PollEvent returns the next event and records key events received during a paste.
func (s *PasteScreen) PollEvent() tcell.Event {
	evt := s.Screen.PollEvent()

	s.mx.Lock()
	defer s.mx.Unlock()
	switch e := evt.(type) {
	case *tcell.EventPaste:
		s.pasting = e.Start()
	case *tcell.EventKey:
		if s.pasting {
			s.pasted[e] = struct{}{}
		}
	}

	return evt
}

// IsPasted reports whether the key event was part of a paste. Each event is reported once.
func (s *PasteScreen) IsPasted(evt *tcell.EventKey) bool {
	s.mx.Lock()
	defer s.mx.Unlock()
	if _, ok := s.pasted[evt]; !ok {
		return false
	}
	delete(s.pasted, evt)

	return true
}

// PastedKey only lets pasted characters through to text inputs. Anything else is dropped,
// so pasted text never triggers hotkeys or confirms a dialog.
func PastedKey(focus tview.Primitive, evt *tcell.EventKey) *tcell.EventKey {
	if evt.Key() != tcell.KeyRune {
		return nil
	}
	switch focus.(type) {
	case *Prompt, *tview.InputField:
		return evt
	default:
		return nil
	}
}
