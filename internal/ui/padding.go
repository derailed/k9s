// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"strings"
	"unicode"

	"github.com/derailed/k9s/internal/render"
)

// MaxyPad tracks uniform column padding.
type MaxyPad []int

// ComputeMaxColumns figures out column max size and necessary padding.
func ComputeMaxColumns(pads MaxyPad, sortColName string, header render.Header, ee render.RowEvents) {
	const colPadding = 1

	for index, h := range header {
		pads[index] = len(h.Name)
		if h.Name == sortColName {
			pads[index] = len(h.Name) + 2
		}
	}

	var row int
	for _, e := range ee {
		for index, field := range e.Row.Fields {
			width := len(field) + colPadding
			if index < len(pads) && width > pads[index] {
				pads[index] = width
			}
		}
		row++
	}
}

// IsASCII checks if table cell has all ascii characters.
func IsASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// Pad a string up to the given length or truncates if greater than length.
func Pad(s string, width int) string {
	if len(s) == width {
		return s
	}
	if len(s) > width {
		return render.Truncate(s, width)
	}
	return s + strings.Repeat(" ", width-len(s))
}
