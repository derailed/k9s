// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
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
	a.GetTable().SetBorderFocusColor(tcell.ColorAliceBlue)
	a.GetTable().SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorAliceBlue).Attributes(tcell.AttrNone))
	a.AddBindKeysFn(a.bindKeys)
	a.SetContextFn(a.aliasContext)

	return &a
}

// Init initializes the view.
func (a *Alias) Init(ctx context.Context) error {
	if err := a.ResourceViewer.Init(ctx); err != nil {
		return err
	}
	a.GetTable().GetModel().SetNamespace(client.NotNamespaced)

	return nil
}

func (a *Alias) aliasContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeyAliases, a.App().command.alias)
}

func (a *Alias) bindKeys(aa *ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, ui.KeyShiftN, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Delete(tcell.KeyCtrlW, tcell.KeyCtrlL)
	aa.Bulk(ui.KeyMap{
		tcell.KeyEnter: ui.NewKeyAction("Goto", a.gotoCmd, true),
		ui.KeyShiftR:   ui.NewKeyAction("Sort Resource", a.GetTable().SortColCmd("RESOURCE", true), false),
		ui.KeyShiftC:   ui.NewKeyAction("Sort Command", a.GetTable().SortColCmd("COMMAND", true), false),
		ui.KeyShiftA:   ui.NewKeyAction("Sort ApiGroup", a.GetTable().SortColCmd("API-GROUP", true), false),
	})
}

func (a *Alias) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.GetTable().CmdBuff().IsActive() {
		return a.GetTable().activateCmd(evt)
	}

	r, _ := a.GetTable().GetSelection()
	if r != 0 {
		s := ui.TrimCell(a.GetTable().SelectTable, r, 1)
		tokens := strings.Split(s, ",")
		a.App().gotoResource(tokens[0], "", true)
		return nil
	}

	return evt
}
