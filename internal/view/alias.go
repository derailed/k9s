package view

import (
	"context"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

const aliasTitle = "Aliases"

// Alias represents a command alias view.
type Alias struct {
	ResourceViewer
}

// NewAlias returns a new alias view.
func NewAlias(gvr client.GVR) ResourceViewer {
	a := Alias{
		ResourceViewer: NewBrowser(gvr),
	}
	a.GetTable().SetColorerFn(render.Alias{}.ColorerFunc())
	a.GetTable().SetBorderFocusColor(tcell.ColorMediumSpringGreen)
	a.GetTable().SetSelectedStyle(tcell.ColorWhite, tcell.ColorMediumSpringGreen, tcell.AttrNone)
	a.SetBindKeysFn(a.bindKeys)
	a.SetContextFn(a.aliasContext)

	return &a
}

func (a *Alias) Init(ctx context.Context) error {
	if err := a.ResourceViewer.Init(ctx); err != nil {
		return err
	}
	a.GetTable().GetModel().SetNamespace("*")
	return nil
}

func (a *Alias) aliasContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeyAliases, a.App().command.alias)
}

func (a *Alias) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, ui.KeyShiftN, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyActions{
		tcell.KeyEnter: ui.NewKeyAction("Goto", a.gotoCmd, true),
		ui.KeyShiftR:   ui.NewKeyAction("Sort Resource", a.GetTable().SortColCmd(0, true), false),
		ui.KeyShiftC:   ui.NewKeyAction("Sort Command", a.GetTable().SortColCmd(1, true), false),
		ui.KeyShiftA:   ui.NewKeyAction("Sort ApiGroup", a.GetTable().SortColCmd(2, true), false),
	})
}

func (a *Alias) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	r, _ := a.GetTable().GetSelection()
	if r != 0 {
		s := ui.TrimCell(a.GetTable().SelectTable, r, 1)
		tokens := strings.Split(s, ",")
		if err := a.App().gotoResource(tokens[0], "", true); err != nil {
			a.App().Flash().Err(err)
		}
		return nil
	}

	if a.GetTable().SearchBuff().IsActive() {
		return a.GetTable().activateCmd(evt)
	}
	return evt
}
