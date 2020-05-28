package view

import (
	"context"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
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
	r.GetTable().SetColorerFn(render.Reference{}.ColorerFunc())
	r.GetTable().SetBorderFocusColor(tcell.ColorMediumSpringGreen)
	r.GetTable().SetSelectedStyle(tcell.ColorWhite, tcell.ColorMediumSpringGreen, tcell.AttrNone)
	r.SetBindKeysFn(r.bindKeys)

	return &r
}

// Init initializes the view.
func (r *Reference) Init(ctx context.Context) error {
	if err := r.ResourceViewer.Init(ctx); err != nil {
		return err
	}
	r.GetTable().GetModel().SetNamespace(client.AllNamespaces)

	return nil
}

func (r *Reference) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Delete(tcell.KeyCtrlW, tcell.KeyCtrlL, tcell.KeyCtrlZ)
	aa.Add(ui.KeyActions{
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
	gvr := ui.TrimCell(r.GetTable().SelectTable, row, 2)

	if err := r.App().gotoResource(client.NewGVR(gvr).R(), path, false); err != nil {
		r.App().Flash().Err(err)
	}

	return evt
}
