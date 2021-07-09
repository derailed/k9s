package model

import (
	"strings"
)

// MaxNavHistory tracks max command NavHistory
const MaxNavHistory = 20

// NavHistory represents a command NavHistory.
type NavHistory struct {
	commands []string
	limit    int
}

// NewNavHistory returns a new instance.
func NewNavHistory(limit int) *NavHistory {
	return &NavHistory{
		limit: limit,
	}
}

// Push adds a new item.
func (h *NavHistory) Push(c string) {
	if c == "" {
		return
	}
	c = strings.ToLower(c)
	if len(h.commands) < h.limit {
		h.commands = append([]string{c}, h.commands...)
		return
	}
	h.commands = append([]string{c}, h.commands[:len(h.commands)-1]...)
}

func (h *NavHistory) Prev() string {
	if len(h.commands) < 2 {
		return ""
	}
	return h.commands[1]
}
