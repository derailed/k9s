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
	e.GetTable().SetSortCol(7, 0, true)

	return &e
}

func (e *Event) bindKeys(aa ui.KeyActions) {
	aa.Delete(tcell.KeyCtrlD, ui.KeyE)
	aa.Add(ui.KeyActions{
		ui.KeyShiftY: ui.NewKeyAction("Sort Type", e.GetTable().SortColCmd(1, true), false),
		ui.KeyShiftR: ui.NewKeyAction("Sort Reason", e.GetTable().SortColCmd(2, true), false),
		ui.KeyShiftE: ui.NewKeyAction("Sort Source", e.GetTable().SortColCmd(3, true), false),
		ui.KeyShiftC: ui.NewKeyAction("Sort Count", e.GetTable().SortColCmd(4, true), false),
	})
}
