// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

// Reference represents resource references.
type Reference struct {
	ResourceViewer
}

// NewReference returns a new alias view.
func NewReference(gvr client.GVR) ResourceViewer {
	r := Reference{
		ResourceViewer: NewBrowser(gvr),
	}
	r.GetTable().SetBorderFocusColor(tcell.ColorMediumSpringGreen)
	r.GetTable().SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorMediumSpringGreen).Attributes(tcell.AttrNone))
	r.AddBindKeysFn(r.bindKeys)

	return &r
}

// Init initializes the view.
func (r *Reference) Init(ctx context.Context) error {
	if err := r.ResourceViewer.Init(ctx); err != nil {
		return err
	}
	r.GetTable().GetModel().SetNamespace(client.BlankNamespace)

	return nil
}

func (r *Reference) bindKeys(aa *ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Delete(tcell.KeyCtrlW, tcell.KeyCtrlL, tcell.KeyCtrlZ)
	aa.Bulk(ui.KeyMap{
		tcell.KeyEnter: ui.NewKeyAction("Goto", r.gotoCmd, true),
		ui.KeyShiftV:   ui.NewKeyAction("Sort GVR", r.GetTable().SortColCmd("GVR", true), false),
	})
}

func (r *Reference) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	row, _ := r.GetTable().GetSelection()
	if row == 0 {
		return evt
	}

	path := r.GetTable().GetSelectedItem()
	ns, _ := client.Namespaced(path)
	gvr := ui.TrimCell(r.GetTable().SelectTable, row, 2)
	r.App().gotoResource(client.NewGVR(gvr).R()+" "+ns, path, false)

	return evt
}
