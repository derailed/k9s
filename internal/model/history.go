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
	commands             []string
	limit                int
	activeCommandIndex   int
	previousCommandIndex int
}

// NewHistory returns a new instance.
func NewHistory(limit int) *History {
	return &History{
		limit: limit,
	}
}

// Last switches the current and previous history index positions so the
// new command referenced by the index is the previous command
func (h *History) Last() bool {
	if h.Empty() {
		return false
	}

	h.activeCommandIndex, h.previousCommandIndex = h.previousCommandIndex, h.activeCommandIndex
	return true
}

// Back moves the history position index back by one
func (h *History) Back() bool {
	if h.Empty() {
		return false
	}

	// Return if there are no more commands left in the backward history
	if h.activeCommandIndex == 0 {
		return false
	}

	h.previousCommandIndex = h.activeCommandIndex
	h.activeCommandIndex = h.activeCommandIndex - 1
	return true
}

// Forward moves the history position index forward by one
func (h *History) Forward() bool {
	if h.Empty() {
		return false
	}

	// Return if there are no more commands left in the forward history
	if h.activeCommandIndex >= len(h.commands)-1 {
		return false
	}

	h.previousCommandIndex = h.activeCommandIndex
	h.activeCommandIndex = h.activeCommandIndex + 1
	return true
}

// CurrentIndex returns the current index of the active command in the history
func (h *History) CurrentIndex() int {
	return h.activeCommandIndex
}

// PreviousIndex returns the index of the command that was the most recent
// active command in the history
func (h *History) PreviousIndex() int {
	return h.previousCommandIndex
}

// Pop removes the single most recent history item
// and returns a bool if the list changed.
func (h *History) Pop() bool {
	return h.PopN(1)
}

// PopN removes the N most recent history item
// and returns a bool if the list changed.
// Argument specifies how many to remove from the history
func (h *History) PopN(n int) bool {
	cmdLength := len(h.commands)
	if cmdLength == 0 {
		return false
	}

	h.commands = h.commands[:cmdLength-n]
	return true
}

// List returns the current command history.
func (h *History) List() []string {
	return h.commands
}

// Push adds a new item.
func (h *History) Push(c string) {
	if c == "" {
		return
	}

	c = strings.ToLower(c)
	if len(h.commands) < h.limit {
		h.commands = append(h.commands, c)
		h.previousCommandIndex = h.activeCommandIndex
		h.activeCommandIndex = len(h.commands) - 1
		return
	}
}

// Clear clears out the stack.
func (h *History) Clear() {
	h.commands = nil
	h.activeCommandIndex = 0
	h.previousCommandIndex = 0
}

// Empty returns true if no history.
func (h *History) Empty() bool {
	return len(h.commands) == 0
}
