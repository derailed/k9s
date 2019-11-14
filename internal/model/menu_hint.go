package model

import (
	"strconv"
	"strings"
)

// MenuHint represents keyboard mnemonic.
type MenuHint struct {
	Mnemonic    string
	Description string
	Visible     bool
}

// IsBlank checks if menu hint is a place holder.
func (m MenuHint) IsBlank() bool {
	return m.Mnemonic == "" && m.Description == "" && m.Visible == false
}

// MenuHints represents a collection of hints.
type MenuHints []MenuHint

func (h MenuHints) Len() int {
	return len(h)
}

func (h MenuHints) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h MenuHints) Less(i, j int) bool {
	n, err1 := strconv.Atoi(h[i].Mnemonic)
	m, err2 := strconv.Atoi(h[j].Mnemonic)
	if err1 == nil && err2 == nil {
		return n < m
	}
	if err1 == nil && err2 != nil {
		return true
	}
	if err1 != nil && err2 == nil {
		return false
	}
	return strings.Compare(h[i].Description, h[j].Description) < 0
}
