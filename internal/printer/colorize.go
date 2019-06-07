package printer

import (
	"fmt"
)

// Color describes a terminal color.
type Color int

// Defines basic ANSI colors.
const (
	Black Color = iota + 30
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

// Colorize a string based on given color.
func Colorize(s string, c Color) string {
	if c == 0 {
		c = White
	}
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", c, s)
}
