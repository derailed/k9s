package view

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
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
func (a *Alias) Init(ctx context.Context) error {
	if err := a.Table.Init(ctx); err != nil {
		return err
	}

	a.SetColorerFn(render.Alias{}.ColorerFunc())
	a.SetBorderFocusColor(tcell.ColorMediumSpringGreen)
	a.SetSelectedStyle(tcell.ColorWhite, tcell.ColorMediumSpringGreen, tcell.AttrNone)
	a.registerActions()
	a.Update(a.hydrate())
	a.resetTitle()

	return nil
}

func (a *Alias) registerActions() {
	a.Actions().Delete(ui.KeyShiftA, ui.KeyShiftN, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
	a.Actions().Add(ui.KeyActions{
		tcell.KeyEnter:  ui.NewKeyAction("Goto Resource", a.gotoCmd, true),
		tcell.KeyEscape: ui.NewKeyAction("Reset", a.resetCmd, false),
		ui.KeySlash:     ui.NewKeyAction("Filter", a.activateCmd, false),
		ui.KeyShiftR:    ui.NewKeyAction("Sort Resource", a.SortColCmd(0, true), false),
		ui.KeyShiftC:    ui.NewKeyAction("Sort Command", a.SortColCmd(1, true), false),
		ui.KeyShiftA:    ui.NewKeyAction("Sort ApiGroup", a.SortColCmd(2, true), false),
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

func (a *Alias) hydrate() render.TableData {
	var re render.Alias

	data := render.TableData{
		Header:    re.Header(render.AllNamespaces),
		RowEvents: make(render.RowEvents, 0, len(aliases.Alias)),
		Namespace: resource.NotNamespaced,
	}

	aa := make(config.ShortNames, len(aliases.Alias))
	for alias, gvr := range aliases.Alias {
		if _, ok := aa[gvr]; ok {
			aa[gvr] = append(aa[gvr], alias)
		} else {
			aa[gvr] = []string{alias}
		}
	}

	for gvr, aliases := range aa {
		var row render.Row
		if err := re.Render(aliases, gvr, &row); err != nil {
			log.Error().Err(err).Msgf("Alias render failed")
			continue
		}
		data.RowEvents = append(data.RowEvents, render.RowEvent{
			Kind: render.EventAdd,
			Row:  row,
		})
	}

	return data
}

func (a *Alias) resetTitle() {
	a.SetTitle(fmt.Sprintf(aliasTitleFmt, aliasTitle, a.GetRowCount()-1))
}
