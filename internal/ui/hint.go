package ui

import (
	"strconv"
	"strings"
)

type (
	// Hint represents keyboard mnemonic.
	Hint struct {
		Mnemonic, Description string
	}
	// Hints a collection of keyboard mnemonics.
	Hints []Hint

	// Hinter returns a collection of mnemonics.
	Hinter interface {
		Hints() Hints
	}
)

func (h Hints) Len() int {
	return len(h)
}

func (h Hints) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h Hints) Less(i, j int) bool {
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
