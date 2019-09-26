package views

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

const (
	aliasTitle    = "Aliases"
	aliasTitleFmt = " [aqua::b]%s([fuchsia::b]%d[fuchsia::-])[aqua::-] "
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
	v.SetBorderFocusColor(tcell.ColorFuchsia)
	v.SetSelectedStyle(tcell.ColorWhite, tcell.ColorFuchsia, tcell.AttrNone)
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

	v.SetActions(ui.KeyActions{
		tcell.KeyEnter:  ui.NewKeyAction("Goto", v.gotoCmd, true),
		tcell.KeyEscape: ui.NewKeyAction("Reset", v.resetCmd, false),
		ui.KeySlash:     ui.NewKeyAction("Filter", v.activateCmd, false),
		ui.KeyShiftR:    ui.NewKeyAction("Sort Resources", v.SortColCmd(1), true),
		ui.KeyShiftO:    ui.NewKeyAction("Sort Groups", v.SortColCmd(2), true),
	})
}

func (v *aliasView) getTitle() string {
	return aliasTitle
}

func (v *aliasView) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.Cmd().Empty() {
		v.Cmd().Reset()
		return nil
	}

	return v.backCmd(evt)
}

func (v *aliasView) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	r, _ := v.GetSelection()
	if r != 0 {
		return v.runCmd(evt)
	}

	if v.Cmd().IsActive() {
		return v.activateCmd(evt)
	}

	return evt
}

func (v *aliasView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cancel != nil {
		v.cancel()
	}

	if v.Cmd().IsActive() {
		v.Cmd().Reset()
	} else {
		v.app.inject(v.current)
	}

	return nil
}

func (v *aliasView) runCmd(evt *tcell.EventKey) *tcell.EventKey {
	r, _ := v.GetSelection()
	if r > 0 {
		v.app.gotoResource(ui.TrimCell(v.Table, r, 0), true)
	}

	return nil
}

func (v *aliasView) hydrate() resource.TableData {
	cmds := make(map[string]*resCmd, 40)
	aliasCmds(v.app.Conn(), cmds)

	data := resource.TableData{
		Header:    resource.Row{"ALIAS", "RESOURCE", "APIGROUP"},
		Rows:      make(resource.RowEvents, len(cmds)),
		Namespace: resource.NotNamespaced,
	}

	for k := range cmds {
		fields := resource.Row{
			ui.Pad(k, 30),
			ui.Pad(cmds[k].gvr, 30),
			ui.Pad(cmds[k].api, 30),
		}
		data.Rows[k] = &resource.RowEvent{
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
