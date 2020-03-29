package color

import (
	"fmt"
)

// ColorFmt colorize a string with ansi colors.
const ColorFmt = "\x1b[%dm%s\x1b[0m"

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
	return fmt.Sprintf(ColorFmt, c, s)
}
