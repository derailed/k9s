package color

import (
	"fmt"
)

// Paint describes a terminal color.
type Paint int

// Defines basic ANSI colors.
const (
	Black Paint = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta
	Cyan
	White
	DarkGray = 90

	Bold = 1
)

// Colorize returns an ASCII colored string based on given color.
func Colorize(s string, c Paint) string {
	if c == 0 {
		c = White
	}
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", c, s)
}
