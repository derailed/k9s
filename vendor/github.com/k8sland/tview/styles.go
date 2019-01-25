package tview

import "github.com/gdamore/tcell"

// Styles defines various colors used when primitives are initialized. These
// may be changed to accommodate a different look and feel.
//
// The default is for applications with a black background and basic colors:
// black, white, yellow, green, and blue.
var Styles = struct {
	PrimitiveBackgroundColor    tcell.Color // Main background color for primitives.
	ContrastBackgroundColor     tcell.Color // Background color for contrasting elements.
	MoreContrastBackgroundColor tcell.Color // Background color for even more contrasting elements.
	BorderColor                 tcell.Color // Box borders.
	TitleColor                  tcell.Color // Box titles.
	FocusColor                  tcell.Color // Box focus color.
	GraphicsColor               tcell.Color // Graphics.
	PrimaryTextColor            tcell.Color // Primary text.
	SecondaryTextColor          tcell.Color // Secondary text (e.g. labels).
	TertiaryTextColor           tcell.Color // Tertiary text (e.g. subtitles, notes).
	InverseTextColor            tcell.Color // Text on primary-colored backgrounds.
	ContrastSecondaryTextColor  tcell.Color // Secondary text on ContrastBackgroundColor-colored backgrounds.
}{
	PrimitiveBackgroundColor:    tcell.ColorBlack,
	ContrastBackgroundColor:     tcell.ColorBlue,
	MoreContrastBackgroundColor: tcell.ColorGreen,
	BorderColor:                 tcell.ColorWhite,
	TitleColor:                  tcell.ColorWhite,
	FocusColor:                  tcell.ColorGreen,
	GraphicsColor:               tcell.ColorWhite,
	PrimaryTextColor:            tcell.ColorWhite,
	SecondaryTextColor:          tcell.ColorYellow,
	TertiaryTextColor:           tcell.ColorGreen,
	InverseTextColor:            tcell.ColorBlue,
	ContrastSecondaryTextColor:  tcell.ColorDarkCyan,
}
