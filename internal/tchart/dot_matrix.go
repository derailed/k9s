package tchart

import (
	"fmt"

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

// Segment represents a dial segment.
type Segment []int

// Segments represents a collection of segments.
type Segments []Segment

// Matrix represents a number dial.
type Matrix [][]rune

// Orientation tracks char orientations.
type Orientation int

// DotMatrix tracks a char matrix.
type DotMatrix struct {
	row, col int
}

// NewDotMatrix returns a new matrix.
func NewDotMatrix(row, col int) DotMatrix {
	return DotMatrix{
		row: row,
		col: col,
	}
}

// Print prints the matrix.
func (d DotMatrix) Print(n int) Matrix {
	if d.row == d.col {
		return To3x3Char(n)
	}
	m := make(Matrix, d.row)
	segs := asSegments(n)
	for row := 0; row < d.row; row++ {
		for col := 0; col < d.col; col++ {
			m[row] = append(m[row], segs.CharFor(row, col))
		}
	}
	return m
}

func asSegments(n int) Segment {
	switch n {
	case 0:
		return Segment{1, 1, 1, 0, 1, 1, 1}
	case 1:
		return Segment{0, 0, 1, 0, 0, 1, 0}
	case 2:
		return Segment{1, 0, 1, 1, 1, 0, 1}
	case 3:
		return Segment{1, 0, 1, 1, 0, 1, 1}
	case 4:
		return Segment{0, 1, 0, 1, 0, 1, 0}
	case 5:
		return Segment{1, 1, 0, 1, 0, 1, 1}
	case 6:
		return Segment{0, 1, 0, 1, 1, 1, 1}
	case 7:
		return Segment{1, 0, 1, 0, 0, 1, 0}
	case 8:
		return Segment{1, 1, 1, 1, 1, 1, 1}
	case 9:
		return Segment{1, 1, 1, 1, 0, 1, 0}

	default:
		panic(fmt.Sprintf("NYI %d", n))
	}
}

// CharFor returns a char based on row/col.
func (s Segment) CharFor(row, col int) rune {
	c := dots[0]
	segs := ToSegments(row, col)
	if segs == nil {
		return c
	}
	for _, seg := range segs {
		if s[seg] == 1 {
			c = charForSeg(seg, row, col)
		}
	}
	return c
}

func charForSeg(seg, row, col int) rune {
	switch seg {
	case 0, 3, 6:
		return dots[2]
	}
	if row == 0 && (col == 0 || col == 2) {
		return dots[2]
	}

	return dots[3]
}

var segs = map[int][][]int{
	0: {{1, 0}, {0}, {2, 0}},
	1: {{1}, nil, {2}},
	2: {{1, 3}, {3}, {2, 3}},
	3: {{4}, nil, {5}},
	4: {{4, 6}, {6}, {5, 6}},
}

// ToSegments return path segments.
func ToSegments(row, col int) []int {
	return segs[row][col]
}

// To3x3Char returns 3x3 number matrix
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
