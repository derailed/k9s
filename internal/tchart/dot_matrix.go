// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tchart

import (
	"github.com/derailed/tview"
)

var dots = []rune{' ', '⠂', '▤', '▥'}

const (
	b    = ' '
	h    = tview.BoxDrawingsHeavyHorizontal
	v    = tview.BoxDrawingsHeavyVertical
	tl   = tview.BoxDrawingsHeavyDownAndRight
	tr   = tview.BoxDrawingsHeavyDownAndLeft
	bl   = tview.BoxDrawingsHeavyUpAndRight
	br   = tview.BoxDrawingsHeavyUpAndLeft
	teeL = tview.BoxDrawingsHeavyVerticalAndLeft
	teeR = tview.BoxDrawingsHeavyVerticalAndRight
	lh   = '\u2578'
	rh   = '\u257a'
	hv   = '\u2579'
	lv   = '\u257b'
)

// Matrix represents a number dial.
type Matrix [][]rune

// Orientation tracks char orientations.
type Orientation int

// DotMatrix tracks a char matrix.
type DotMatrix struct {
	row, col int
}

// NewDotMatrix returns a new matrix.
func NewDotMatrix() DotMatrix {
	return DotMatrix{
		row: 3,
		col: 3,
	}
}

// Print prints the matrix.
func (d DotMatrix) Print(n int) Matrix {
	return To3x3Char(n)
}

// To3x3Char returns 3x3 number matrix.
func To3x3Char(numb int) Matrix {
	switch numb {
	case 1:
		return Matrix{
			[]rune{b, lv, b},
			[]rune{b, v, b},
			[]rune{b, hv, b},
		}
	case 2:
		return Matrix{
			[]rune{rh, h, tr},
			[]rune{tl, h, br},
			[]rune{bl, h, lh},
		}
	case 3:
		return Matrix{
			[]rune{h, h, tr},
			[]rune{rh, h, teeL},
			[]rune{h, h, br},
		}
	case 4:
		return Matrix{
			[]rune{lv, b, lv},
			[]rune{bl, h, teeL},
			[]rune{b, b, hv},
		}
	case 5:
		return Matrix{
			[]rune{tl, h, lh},
			[]rune{bl, h, tr},
			[]rune{rh, h, br},
		}
	case 6:
		return Matrix{
			[]rune{tl, h, lh},
			[]rune{teeR, h, tr},
			[]rune{bl, h, br},
		}
	case 7:
		return Matrix{
			[]rune{h, h, tr},
			[]rune{b, b, v},
			[]rune{b, b, hv},
		}
	case 8:
		return Matrix{
			[]rune{tl, h, tr},
			[]rune{teeR, h, teeL},
			[]rune{bl, h, br},
		}
	case 9:
		return Matrix{
			[]rune{tl, h, tr},
			[]rune{bl, h, teeL},
			[]rune{rh, h, br},
		}
	default:
		return Matrix{
			[]rune{tl, h, tr},
			[]rune{v, b, v},
			[]rune{bl, h, br},
		}
	}
}
