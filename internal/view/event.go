// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

// Event represents a command alias view.
type Event struct {
	ResourceViewer
}

// NewEvent returns a new alias view.
func NewEvent(gvr *client.GVR) ResourceViewer {
	e := Event{
		ResourceViewer: NewBrowser(gvr),
	}
	e.AddBindKeysFn(e.bindKeys)
	e.GetTable().SetSortCol("LAST SEEN", false)

	return &e
}

func (e *Event) bindKeys(aa *ui.KeyActions) {
	aa.Delete(tcell.KeyCtrlD, ui.KeyE, ui.KeyA)
	aa.Bulk(ui.KeyMap{
		tcell.KeyCtrlZ: ui.NewKeyAction("Toggle Warnings/Errors", e.toggleWarnErrorCmd, false),
		ui.KeyShiftL:   ui.NewKeyAction("Sort LastSeen", e.GetTable().SortColCmd("LAST SEEN", false), false),
		ui.KeyShiftF:   ui.NewKeyAction("Sort FirstSeen", e.GetTable().SortColCmd("FIRST SEEN", false), false),
		ui.KeyShiftT:   ui.NewKeyAction("Sort Type", e.GetTable().SortColCmd("TYPE", true), false),
		ui.KeyShiftR:   ui.NewKeyAction("Sort Reason", e.GetTable().SortColCmd("REASON", true), false),
		ui.KeyShiftS:   ui.NewKeyAction("Sort Source", e.GetTable().SortColCmd("SOURCE", true), false),
		ui.KeyShiftC:   ui.NewKeyAction("Sort Count", e.GetTable().SortColCmd("COUNT", true), false),
	})
}

func (e *Event) toggleWarnErrorCmd(*tcell.EventKey) *tcell.EventKey {
	b, ok := e.ResourceViewer.(*Browser)
	if !ok {
		return nil
	}
	filter := b.CmdBuff().GetText()
	if filter == "Warning|Error" {
		e.SetFilter("", true)
		e.App().Flash().Info("Showing all events")
	} else {
		e.SetFilter("Warning|Error", true)
		e.App().Flash().Info("Showing Warning and Error events only")
	}
	return nil
}
