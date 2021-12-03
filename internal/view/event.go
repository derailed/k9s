package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell/v2"
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
	e.AddBindKeysFn(e.bindKeys)
	e.GetTable().SetSortCol("LAST SEEN", false)

	return &e
}

func (e *Event) bindKeys(aa ui.KeyActions) {
	aa.Delete(tcell.KeyCtrlD, ui.KeyE)
	aa.Add(ui.KeyActions{
		ui.KeyShiftL: ui.NewKeyAction("Sort LastSeen", e.GetTable().SortColCmd("LAST SEEN", false), false),
		ui.KeyShiftT: ui.NewKeyAction("Sort Type", e.GetTable().SortColCmd("TYPE", true), false),
		ui.KeyShiftR: ui.NewKeyAction("Sort Reason", e.GetTable().SortColCmd("REASON", true), false),
		ui.KeyShiftS: ui.NewKeyAction("Sort Source", e.GetTable().SortColCmd("SOURCE", true), false),
		ui.KeyShiftC: ui.NewKeyAction("Sort Count", e.GetTable().SortColCmd("COUNT", true), false),
	})
}
