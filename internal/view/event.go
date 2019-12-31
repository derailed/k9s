package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

// Event represents a command alias view.
type Event struct {
	ResourceViewer
}

// NewEvent returns a new alias view.
func NewEvent(gvr client.GVR) ResourceViewer {
	e := Event{
		ResourceViewer: NewBrowser(gvr),
	}
	e.GetTable().SetColorerFn(render.Event{}.ColorerFunc())
	e.SetBindKeysFn(e.bindKeys)

	return &e
}

func (e *Event) bindKeys(aa ui.KeyActions) {
	aa.Delete(tcell.KeyCtrlD, ui.KeyE)
}
