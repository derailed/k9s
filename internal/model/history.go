// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"strings"
)

// MaxHistory tracks max command history.
const MaxHistory = 20

// CommandState represents a complete command state including filters.
type CommandState struct {
	Command string `yaml:"command"`
	Filter  string `yaml:"filter,omitempty"`
	Labels  string `yaml:"labels,omitempty"`
}

// History represents a command history.
type History struct {
	states     []*CommandState
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
	commands := make([]string, len(h.states))
	for i, state := range h.states {
		if state != nil {
			commands[i] = state.Command
		}
	}
	return commands
}

// Top returns the last command in the history if present.
func (h *History) Top() (*CommandState, bool) {
	h.currentIdx = len(h.states) - 1
	return h.at(h.currentIdx)
}

// Last returns the nth command prior to last.
func (h *History) Last(idx int) (*CommandState, bool) {
	h.currentIdx = len(h.states) - idx
	return h.at(h.currentIdx)
}

func (h *History) at(idx int) (*CommandState, bool) {
	if idx < 0 || idx >= len(h.states) {
		return nil, false
	}

	state := h.states[idx]
	if state == nil || state.IsEmpty() {
		return nil, false
	}

	return state, true
}

// Back moves the history position index back by one.
func (h *History) Back() (*CommandState, bool) {
	if h.Empty() || h.currentIdx <= 0 {
		return nil, false
	}
	h.currentIdx--
	return h.at(h.currentIdx)
}

// Forward moves the history position index forward by one
func (h *History) Forward() (*CommandState, bool) {
	h.currentIdx++
	if h.Empty() || h.currentIdx >= len(h.states) {
		return nil, false
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
	pop := len(h.states) - n
	if h.Empty() || pop < 0 {
		return false
	}
	h.states = h.states[:pop]
	h.currentIdx = len(h.states) - 1

	return true
}

// Push adds a new command or command state to the history.
// Accepts either a string or *CommandState.
func (h *History) Push(cmd interface{}) {
	var state *CommandState

	switch v := cmd.(type) {
	case string:
		state = NewCommandState(strings.ToLower(v), "", "")
	case *CommandState:
		state = v
	default:
		return
	}

	if state == nil || state.IsEmpty() || len(h.states) >= h.limit {
		return
	}
	if h.currentIdx < len(h.states)-1 {
		h.states = h.states[:h.currentIdx+1]
	}
	h.states = append(h.states, state)
	h.currentIdx = len(h.states) - 1
}

// Clear clears out the stack.
func (h *History) Clear() {
	h.states = nil
	h.currentIdx = -1
}

// Empty returns true if no history.
func (h *History) Empty() bool {
	return len(h.states) == 0
}

// States returns the command history with full state information.
func (h *History) States() []*CommandState {
	return h.states
}

func NewCommandState(command, filter, labels string) *CommandState {
	return &CommandState{
		Command: command,
		Filter:  filter,
		Labels:  labels,
	}
}

func (cs *CommandState) IsEmpty() bool {
	return cs.Command == ""
}

func (cs *CommandState) HasFilters() bool {
	return cs.Filter != "" || cs.Labels != ""
}

func (cs *CommandState) String() string {
	cmd := cs.Command
	if cs.Filter != "" {
		cmd += " /" + cs.Filter
	}
	if cs.Labels != "" {
		cmd += " " + cs.Labels
	}
	return cmd
}
