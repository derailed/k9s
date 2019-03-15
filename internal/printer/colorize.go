package printer

import (
	"fmt"
)

// Defines basic ANSI colors.
const (
	ColorBlack = iota + 30
	ColorRed
	ColorGreen
	ColorYellow
	ColorBlue
	ColorMagenta
	ColorCyan
	ColorWhite

	ColorBold     = 1
	ColorDarkGray = 90
)

// Colorize a string based on given color.
func Colorize(s string, c int) string {
	return fmt.Sprintf("\x1b[%dm%s\x1b[0m", c, s)
}
