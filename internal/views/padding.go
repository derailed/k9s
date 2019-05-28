package views

import (
	"strings"
	"time"
	"unicode"

	"github.com/derailed/k9s/internal/resource"
	"k8s.io/apimachinery/pkg/util/duration"
)

type maxyPad []int

func computeMaxColumns(pads maxyPad, sortCol int, table resource.TableData) {
	const colPadding = 1

	for index, h := range table.Header {
		pads[index] = len(h)
		if index == sortCol {
			pads[index] = len(h) + 2
		}
	}

	var row int
	for k, rev := range table.Rows {
		ageIndex := len(rev.Fields) - 1
		for index, field := range rev.Fields {
			// Date field comes out as timestamp.
			if index == ageIndex {
				dur, err := time.ParseDuration(field)
				if err == nil {
					field = duration.HumanDuration(dur)
				}
				table.Rows[k].Fields[index] = field
			}
			width := len(field) + colPadding
			if width > pads[index] {
				pads[index] = width
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
