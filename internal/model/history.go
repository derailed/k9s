package model

import (
	"strings"

	"github.com/rs/zerolog/log"
)

// MaxHistory tracks max command history
const MaxHistory = 20

// History represents a command history.
type History struct {
	commands []string
	limit    int
}

// NewHistory returns a new instance.
func NewHistory(limit int) *History {
	return &History{
		limit: limit,
	}
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
	if i := h.indexOf(c); i != -1 {
		return
	}
	if len(h.commands) < h.limit {
		h.commands = append([]string{c}, h.commands...)
		return
	}
	h.commands = append([]string{c}, h.commands[:len(h.commands)-1]...)
}

// Clear clears out the stack.
func (h *History) Clear() {
	log.Debug().Msgf("History CLEARED!!!")
	h.commands = nil
}

// Empty returns true if no history.
func (h *History) Empty() bool {
	return len(h.commands) == 0
}

func (h *History) indexOf(s string) int {
	for i, c := range h.commands {
		if c == s {
			return i
		}
	}
	return -1
}
