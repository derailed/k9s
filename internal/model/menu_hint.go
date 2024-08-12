// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

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
	return m.Mnemonic == "" && m.Description == "" && !m.Visible
}

// String returns a string representation.
func (m MenuHint) String() string {
	return m.Mnemonic
}

// MenuHints represents a collection of hints.
type MenuHints []MenuHint

// Len returns the hints length.
func (h MenuHints) Len() int {
	return len(h)
}

// Swap swaps to elements.
func (h MenuHints) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

// Less returns true if first hint is less than second.
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
