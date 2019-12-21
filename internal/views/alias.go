package views

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

type aliasView struct {
	*tableView

	app     *appView
	current ui.Igniter
	cancel  context.CancelFunc
}

func newAliasView(app *appView, current ui.Igniter) *aliasView {
	v := aliasView{
		tableView: newTableView(app, aliasTitle),
		app:       app,
	}
	v.SetBorderFocusColor(tcell.ColorMediumSpringGreen)
	v.SetSelectedStyle(tcell.ColorWhite, tcell.ColorMediumSpringGreen, tcell.AttrNone)
	v.SetColorerFn(aliasColorer)
	v.current = current
	v.SetActiveNS("")
	v.registerActions()

	return &v
}

// Init the view.
func (v *aliasView) Init(context.Context, string) {
	v.Update(v.hydrate())
	v.app.SetFocus(v)
	v.resetTitle()
	v.app.SetHints(v.Hints())
}

func (v *aliasView) registerActions() {
	v.RmAction(ui.KeyShiftA)
	v.RmAction(ui.KeyShiftN)
	v.RmAction(tcell.KeyCtrlS)

	v.SetActions(ui.KeyActions{
		tcell.KeyEnter:  ui.NewKeyAction("Goto", v.gotoCmd, true),
		tcell.KeyEscape: ui.NewKeyAction("Reset", v.resetCmd, false),
		ui.KeySlash:     ui.NewKeyAction("Filter", v.activateCmd, false),
		ui.KeyShiftR:    ui.NewKeyAction("Sort Resource", ui.SortColCmd(v, 0, true), false),
		ui.KeyShiftC:    ui.NewKeyAction("Sort Command", ui.SortColCmd(v, 1, true), false),
		ui.KeyShiftA:    ui.NewKeyAction("Sort ApiGroup", ui.SortColCmd(v, 2, true), false),
	})
}

func (v *aliasView) getTitle() string {
	return aliasTitle
}

func (v *aliasView) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.SearchBuff().Empty() {
		v.SearchBuff().Reset()
		return nil
	}

	return v.backCmd(evt)
}

func (v *aliasView) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	r, _ := v.GetSelection()
	if r != 0 {
		s := ui.TrimCell(v.Table, r, 1)
		tokens := strings.Split(s, ",")
		v.app.gotoResource(tokens[0], true)
		return nil
	}

	if v.SearchBuff().IsActive() {
		return v.activateCmd(evt)
	}
	return evt
}

func (v *aliasView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cancel != nil {
		v.cancel()
	}

	if v.SearchBuff().IsActive() {
		v.SearchBuff().Reset()
	} else {
		v.app.inject(v.current)
	}

	return nil
}

func (v *aliasView) hydrate() resource.TableData {
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

func (v *aliasView) resetTitle() {
	v.SetTitle(fmt.Sprintf(aliasTitleFmt, aliasTitle, v.GetRowCount()-1))
}
