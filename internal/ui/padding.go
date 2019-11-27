package ui

import (
	"strings"
	"unicode"

	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/resource"
)

// MaxyPad tracks uniform column padding.
type MaxyPad []int

// ComputeMaxColumns figures out column max size and necessary padding.
func ComputeMaxColumns(pads MaxyPad, sortCol int, header render.HeaderRow, ee render.RowEvents) {
	const colPadding = 1

	for index, h := range header {
		pads[index] = len(h.Name)
		if index == sortCol {
			pads[index] = len(h.Name) + 2
		}
	}

	var row int
	for _, e := range ee {
		for index, field := range e.Row.Fields {
			width := len(field) + colPadding
			if width > pads[index] {
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
		return resource.Truncate(s, width)
	}

	return s + strings.Repeat(" ", width-len(s))
}
