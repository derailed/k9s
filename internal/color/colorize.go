// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package color

import (
	"fmt"
	"strconv"
)

const colorFmt = "\x1b[%dm%s\x1b[0m"

// Paint describes a terminal color.
type Paint int

// Defines basic ANSI colors.
const (
	Black     Paint = iota + 30 // 30
	Red                         // 31
	Green                       // 32
	Yellow                      // 33
	Blue                        // 34
	Magenta                     // 35
	Cyan                        // 36
	LightGray                   // 37
	DarkGray  = 90

	Bold = 1
)

// Colorize returns an ASCII colored string based on given color.
func Colorize(s string, c Paint) string {
	if c == 0 {
		return s
	}
	return fmt.Sprintf(colorFmt, c, s)
}

// ANSIColorize colors a string.
func ANSIColorize(text string, color int) string {
	return "\033[38;5;" + strconv.Itoa(color) + "m" + text + "\033[0m"
}

// Highlight colorize bytes at given indices.
func Highlight(bb []byte, ii []int, c int) []byte {
	b := make([]byte, 0, len(bb))
	for i, j := 0, 0; i < len(bb); i++ {
		if j < len(ii) && ii[j] == i {
			b = append(b, colorizeByte(bb[i], c)...)
			j++
		} else {
			b = append(b, bb[i])
		}
	}

	return b
}

func colorizeByte(b byte, color int) []byte {
	return []byte(ANSIColorize(string(b), color))
}
