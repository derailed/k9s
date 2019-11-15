package view

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

const (
	aliasTitle    = "Aliases"
	aliasTitleFmt = " [mediumseagreen::b]%s([fuchsia::b]%d[fuchsia::-][mediumseagreen::-]) "
)

// Alias represents a command alias view.
type Alias struct {
	*Table
}

// NewAlias returns a new alias view.
func NewAlias() *Alias {
	return &Alias{
		Table: NewTable(aliasTitle),
	}
}

// Init the view.
func (a *Alias) Init(ctx context.Context) {
	a.Table.Init(ctx)

	a.SetBorderFocusColor(tcell.ColorMediumSpringGreen)
	a.SetSelectedStyle(tcell.ColorWhite, tcell.ColorMediumSpringGreen, tcell.AttrNone)
	a.SetColorerFn(aliasColorer)
	a.ActiveNS = resource.AllNamespaces
	a.registerActions()

	a.Update(a.hydrate())
	a.resetTitle()
}

func (a *Alias) Name() string {
	return aliasTitle
}

func (a *Alias) Start() {}
func (a *Alias) Stop()  {}

func (a *Alias) registerActions() {
	a.RmAction(ui.KeyShiftA)
	a.RmAction(ui.KeyShiftN)
	a.RmAction(tcell.KeyCtrlS)
	a.RmAction(tcell.KeyCtrlSpace)
	a.RmAction(ui.KeySpace)

	a.AddActions(ui.KeyActions{
		tcell.KeyEnter:  ui.NewKeyAction("Goto Resource", a.gotoCmd, true),
		tcell.KeyEscape: ui.NewKeyAction("Reset", a.resetCmd, false),
		ui.KeySlash:     ui.NewKeyAction("Filter", a.activateCmd, false),
		ui.KeyShiftR:    ui.NewKeyAction("Sort Resource", a.SortColCmd(0), false),
		ui.KeyShiftC:    ui.NewKeyAction("Sort Command", a.SortColCmd(1), false),
		ui.KeyShiftA:    ui.NewKeyAction("Sort ApiGroup", a.SortColCmd(2), false),
	})
}

func (a *Alias) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !a.SearchBuff().Empty() {
		a.SearchBuff().Reset()
		return nil
	}

	return a.backCmd(evt)
}

func (a *Alias) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	r, _ := a.GetSelection()
	if r != 0 {
		s := ui.TrimCell(a.Table.SelectTable, r, 1)
		tokens := strings.Split(s, ",")
		a.app.Content.Pop()
		if !a.app.gotoResource(tokens[0]) {
			a.app.Flash().Err(fmt.Errorf("Goto %s failed", tokens[0]))
		}
		return nil
	}

	if a.SearchBuff().IsActive() {
		return a.activateCmd(evt)
	}

	return evt
}

func (a *Alias) backCmd(_ *tcell.EventKey) *tcell.EventKey {
	if a.SearchBuff().IsActive() {
		a.SearchBuff().Reset()
	} else {
		a.app.Content.Pop()
	}

	return nil
}

func (a *Alias) hydrate() resource.TableData {
	data := resource.TableData{
		Header:    resource.Row{"RESOURCE", "COMMAND", "APIGROUP"},
		Rows:      make(resource.RowEvents, len(aliases.Alias)),
		Namespace: resource.NotNamespaced,
	}

	aa := make(map[string][]string, len(aliases.Alias))
	for alias, gvr := range aliases.Alias {
		if _, ok := aa[gvr]; ok {
			aa[gvr] = append(aa[gvr], alias)
		} else {
			aa[gvr] = []string{alias}
		}
	}

	for gvr, aliases := range aa {
		g := k8s.GVR(gvr)
		fields := resource.Row{
			ui.Pad(g.ToR(), 30),
			ui.Pad(strings.Join(aliases, ","), 70),
			ui.Pad(g.ToG(), 30),
		}
		data.Rows[string(gvr)] = &resource.RowEvent{
			Action: resource.New,
			Fields: fields,
			Deltas: fields,
		}
	}

	return data
}

func (a *Alias) resetTitle() {
	a.SetTitle(fmt.Sprintf(aliasTitleFmt, aliasTitle, a.GetRowCount()-1))
}
