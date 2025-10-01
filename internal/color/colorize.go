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
	if len(ii) == 0 {
		return bb
	}

	result := make([]byte, 0, len(bb)+len(ii)*20) // Extra space for color codes

	// Create a map of byte positions that should be highlighted
	highlightMap := make(map[int]bool)
	for _, pos := range ii {
		highlightMap[pos] = true
	}

	// Process each byte
	for i := 0; i < len(bb); i++ {
		if highlightMap[i] {
			// Check if this is the start of a UTF-8 character
			if (bb[i] & 0xC0) != 0x80 {
				// This is the start of a character, find the end
				charStart := i
				charEnd := i + 1
				for charEnd < len(bb) && (bb[charEnd]&0xC0) == 0x80 {
					charEnd++
				}
				// Colorize the entire character
				char := string(bb[charStart:charEnd])
				colored := ANSIColorize(char, c)
				result = append(result, []byte(colored)...)
				i = charEnd - 1 // Skip the rest of the character bytes
			} else {
				// This is a continuation byte, skip it (already handled)
				continue
			}
		} else {
			result = append(result, bb[i])
		}
	}

	return result
}
