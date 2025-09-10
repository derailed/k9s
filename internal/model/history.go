// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"strings"
)

// MaxHistory tracks max command history.
const MaxHistory = 20

// History represents a command history.
type History struct {
	commands   []string
	limit      int
	currentIdx int
}

// NewHistory returns a new instance.
func NewHistory(limit int) *History {
	return &History{
		limit:      limit,
		currentIdx: -1,
	}
}

// List returns the command history.
func (h *History) List() []string {
	return h.commands
}

// Top returns the last command in the history if present.
func (h *History) Top() (string, bool) {
	h.currentIdx = len(h.commands) - 1

	return h.at(h.currentIdx)
}

// Last returns the nth command prior to last.
func (h *History) Last(idx int) (string, bool) {
	h.currentIdx = len(h.commands) - idx

	return h.at(h.currentIdx)
}

func (h *History) at(idx int) (string, bool) {
	if idx < 0 || idx >= len(h.commands) {
		return "", false
	}

	return h.commands[idx], true
}

// Back moves the history position index back by one.
func (h *History) Back() (string, bool) {
	if h.Empty() || h.currentIdx <= 0 {
		return "", false
	}
	h.currentIdx--

	return h.at(h.currentIdx)
}

// Forward moves the history position index forward by one
func (h *History) Forward() (string, bool) {
	h.currentIdx++
	if h.Empty() || h.currentIdx >= len(h.commands) {
		return "", false
	}

	return h.at(h.currentIdx)
}

// Pop removes the single most recent history item
// and returns a bool if the list changed.
func (h *History) Pop() bool {
	return h.popN(1)
}

// PopN removes the N most recent history item
// and returns a bool if the list changed.
// Argument specifies how many to remove from the history
func (h *History) popN(n int) bool {
	pop := len(h.commands) - n
	if h.Empty() || pop < 0 {
		return false
	}
	h.commands = h.commands[:pop]
	h.currentIdx = len(h.commands) - 1

	return true
}

// Push adds a new item.
func (h *History) Push(c string) {
	if c == "" || len(h.commands) >= h.limit {
		return
	}
	if h.currentIdx < len(h.commands)-1 {
		h.commands = h.commands[:h.currentIdx+1]
	}
	h.commands = append(h.commands, strings.ToLower(c))
	h.currentIdx = len(h.commands) - 1
}

// Clear clears out the stack.
func (h *History) Clear() {
	h.commands = nil
	h.currentIdx = -1
}

// Empty returns true if no history.
func (h *History) Empty() bool {
	return len(h.commands) == 0
}
