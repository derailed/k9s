package tchart_test

import (
	"strconv"
	"testing"

	"github.com/derailed/k9s/internal/tchart"
	"github.com/stretchr/testify/assert"
)

func TestSegmentFor(t *testing.T) {
	uu := map[string]struct {
		r, c int
		e    []int
	}{
		"0x0": {r: 0, c: 0, e: []int{1, 0}},
		"0x1": {r: 0, c: 1, e: []int{0}},
		"0x2": {r: 0, c: 2, e: []int{2, 0}},
		"1x0": {r: 1, c: 0, e: []int{1}},
		"1x1": {r: 1, c: 1, e: nil},
		"1x2": {r: 1, c: 2, e: []int{2}},
		"2x0": {r: 2, c: 0, e: []int{1, 3}},
		"2x1": {r: 2, c: 1, e: []int{3}},
		"2x2": {r: 2, c: 2, e: []int{2, 3}},
		"3x0": {r: 3, c: 0, e: []int{4}},
		"3x1": {r: 3, c: 1, e: nil},
		"3x2": {r: 3, c: 2, e: []int{5}},
		"4x0": {r: 4, c: 0, e: []int{4, 6}},
		"4x1": {r: 4, c: 1, e: []int{6}},
		"4x2": {r: 4, c: 2, e: []int{5, 6}},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, tchart.ToSegments(u.r, u.c))
		})
	}
}

func TestDial(t *testing.T) {
	d := tchart.NewDotMatrix(5, 3)
	for n := 0; n <= 9; n++ {
		i := n
		t.Run(strconv.Itoa(n), func(t *testing.T) {
			assert.Equal(t, numbers[i], d.Print(i))
		})
	}
}

// Helpers...

const hChar, vChar = '▤', '▥'

var numbers = []tchart.Matrix{
	[][]rune{
		{hChar, hChar, hChar},
		{vChar, ' ', vChar},
		{vChar, ' ', vChar},
		{vChar, ' ', vChar},
		{hChar, hChar, hChar},
	},
	[][]rune{
		{' ', ' ', hChar},
		{' ', ' ', vChar},
		{' ', ' ', vChar},
		{' ', ' ', vChar},
		{' ', ' ', vChar},
	},
	[][]rune{
		{hChar, hChar, hChar},
		{' ', ' ', vChar},
		{hChar, hChar, hChar},
		{vChar, ' ', ' '},
		{hChar, hChar, hChar},
	},
	[][]rune{
		{hChar, hChar, hChar},
		{' ', ' ', vChar},
		{hChar, hChar, hChar},
		{' ', ' ', vChar},
		{hChar, hChar, hChar},
	},
	[][]rune{
		{hChar, ' ', ' '},
		{vChar, ' ', ' '},
		{hChar, hChar, hChar},
		{' ', ' ', vChar},
		{' ', ' ', vChar},
	},
	[][]rune{
		{hChar, hChar, hChar},
		{vChar, ' ', ' '},
		{hChar, hChar, hChar},
		{' ', ' ', vChar},
		{hChar, hChar, hChar},
	},
	[][]rune{
		{hChar, ' ', ' '},
		{vChar, ' ', ' '},
		{hChar, hChar, hChar},
		{vChar, ' ', vChar},
		{hChar, hChar, hChar},
	},
	[][]rune{
		{hChar, hChar, hChar},
		{' ', ' ', vChar},
		{' ', ' ', vChar},
		{' ', ' ', vChar},
		{' ', ' ', vChar},
	},
	[][]rune{
		{hChar, hChar, hChar},
		{vChar, ' ', vChar},
		{hChar, hChar, hChar},
		{vChar, ' ', vChar},
		{hChar, hChar, hChar},
	},
	[][]rune{
		{hChar, hChar, hChar},
		{vChar, ' ', vChar},
		{hChar, hChar, hChar},
		{' ', ' ', vChar},
		{' ', ' ', vChar},
	},
}
