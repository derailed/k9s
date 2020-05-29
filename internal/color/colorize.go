package color

import (
	"fmt"
)

const (
	colorFmt     = "\x1b[%dm%s\x1b[0m"
	ansiColorFmt = "\033[38;5;%dm%s\033[0m"
)

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

// AnsiColorize colors a string.
func AnsiColorize(s string, c int) string {
	return fmt.Sprintf(ansiColorFmt, c, s)
}

// Highlight colorize bytes at given indices.
func Highlight(bb []byte, ii []int, c int) []byte {
	b := make([]byte, 0, len(bb))
	for i, j := 0, 0; i < len(bb); i++ {
		if j < len(ii) && ii[j] == i {
			b = append(b, colorizeByte(bb[i], 209)...)
			j++
		} else {
			b = append(b, bb[i])
		}
	}

	return b
}

func colorizeByte(b byte, color int) []byte {
	return []byte(AnsiColorize(string(b), color))
}
