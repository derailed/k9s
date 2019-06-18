package views

import (
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

type (
	keyHandler interface {
		keyboard(evt *tcell.EventKey) *tcell.EventKey
	}

	actionsFn func(keyActions)

	shellView struct {
		*tview.Application
		configurator

		actions keyActions
		pages   *tview.Pages
		content *tview.Pages
		views   map[string]tview.Primitive
	}
)

func newShellView() *shellView {
	return &shellView{
		Application: tview.NewApplication(),
		actions:     make(keyActions),
		pages:       tview.NewPages(),
		content:     tview.NewPages(),
		views:       make(map[string]tview.Primitive),
	}
}
