package views

import (
	"strings"
	"unicode"

	"github.com/derailed/k9s/internal/resource"
)

type maxyPad []int

func computeMaxColumns(pads maxyPad, sortCol int, table resource.TableData) {
	for index, h := range table.Header {
		pads[index] = len(h)
		if index == sortCol {
			pads[index] = len(h) + 2
		}
	}

	var row int
	for _, rev := range table.Rows {
		for index, field := range rev.Fields {
			if len(field) > pads[index] {
				pads[index] = len([]rune(field))
			}
		}
		row++
	}
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}
	return true
}

// Pad a string up to the given length or truncates if greater than length.
func pad(s string, width int) string {
	if len(s) == width {
		return s
	}

	if len(s) > width {
		return resource.Truncate(s, width)
	}

	return s + strings.Repeat(" ", width-len(s))
}
