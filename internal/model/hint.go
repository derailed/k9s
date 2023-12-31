// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

// HintListener represents a menu hints listener.
type HintListener interface {
	HintsChanged(MenuHints)
}

// Hint represent a hint model.
type Hint struct {
	data      MenuHints
	listeners []HintListener
}

// NewHint return new hint model.
func NewHint() *Hint {
	return &Hint{}
}

// RemoveListener deletes a listener.
func (h *Hint) RemoveListener(l HintListener) {
	victim := -1
	for i, lis := range h.listeners {
		if lis == l {
			victim = i
			break
		}
	}
	if victim == -1 {
		return
	}
	h.listeners = append(h.listeners[:victim], h.listeners[victim+1:]...)
}

// AddListener adds a hint listener.
func (h *Hint) AddListener(l HintListener) {
	h.listeners = append(h.listeners, l)
}

// SetHints set model hints.
func (h *Hint) SetHints(hh MenuHints) {
	h.data = hh
	h.fireChanged()
}

// Peek returns the model data.
func (h *Hint) Peek() MenuHints {
	return h.data
}

func (h *Hint) fireChanged() {
	for _, l := range h.listeners {
		l.HintsChanged(h.data)
	}
}
