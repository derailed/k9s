package views

import (
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

type styles struct {
	color tcell.Color
	attrs tcell.AttrMask
	align int
}

func stylesFor(app *appView, res string, col int) styles {
	switch res {
	case "pod":
		return podStyles(app, col)
	default:
		return defaultStyles(app, col)
	}
}

func podStyles(app *appView, col int) styles {
	st := styles{
		color: ui.StdColor,
		attrs: tcell.AttrReverse,
		align: tview.AlignLeft,
	}

	switch col {
	case 5, 6, 7, 8:
		st.align = tview.AlignLeft
		st.color = tcell.ColorGreen
	}

	return st
}

func defaultStyles(app *appView, col int) styles {
	return styles{
		color: tcell.ColorRed,
		attrs: tcell.AttrReverse,
		align: tview.AlignLeft,
	}
}
