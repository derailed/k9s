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
type ColorerFunc func(ns string, h Header, re RowEvent) tcell.Color

// DefaultColorer set the default table row colors.
func DefaultColorer(ns string, h Header, re RowEvent) tcell.Color {
	if !Happy(ns, h, re.Row) {
		return ErrColor
	}

	switch re.Kind {
	case EventAdd:
		return AddColor
	case EventUpdate:
		return ModColor
	case EventDelete:
		return KillColor
	default:
		return StdColor
	}
}
