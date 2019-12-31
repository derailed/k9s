package render

import "github.com/gdamore/tcell"

var (
	// ModColor row modified color.
	ModColor tcell.Color
	// AddColor row added color.
	AddColor tcell.Color
	// ErrColor row err color.
	ErrColor tcell.Color
	// StdColor row default color.
	StdColor tcell.Color
	// HighlightColor row highlight color.
	HighlightColor tcell.Color
	// KillColor row deleted color.
	KillColor tcell.Color
	// CompletedColor row completed color.
	CompletedColor tcell.Color
)

// ColorerFunc represents a resource row colorer.
type ColorerFunc func(ns string, evt RowEvent) tcell.Color

// DefaultColorer set the default table row colors.
func DefaultColorer(ns string, evt RowEvent) tcell.Color {
	var col = StdColor
	switch evt.Kind {
	case EventAdd:
		col = AddColor
	case EventUpdate:
		col = ModColor
	case EventDelete:
		col = KillColor
	}

	return col
}
