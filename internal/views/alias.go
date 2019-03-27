package views

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
)

const (
	aliasTitle    = "Aliases"
	aliasTitleFmt = " [aqua::b]%s[[aqua::b]%d[aqua::-]][aqua::-] "
)

type aliasView struct {
	*tableView

	current igniter
	cancel  context.CancelFunc
}

func newAliasView(app *appView) *aliasView {
	v := aliasView{tableView: newTableView(app, aliasTitle)}
	{
		v.SetSelectedStyle(tcell.ColorWhite, tcell.ColorFuchsia, tcell.AttrNone)
		v.colorerFn = aliasColorer
		v.current = app.content.GetPrimitive("main").(igniter)
		v.currentNS = ""
		v.registerActions()
	}
	return &v
}

// Init the view.
func (v *aliasView) init(context.Context, string) {
	v.update(v.hydrate())
	v.app.SetFocus(v)
	v.resetTitle()
}

func (v *aliasView) registerActions() {
	v.actions[tcell.KeyEnter] = newKeyAction("Goto", v.gotoCmd, true)
	v.actions[tcell.KeyEscape] = newKeyAction("Reset", v.resetCmd, false)
	v.actions[KeySlash] = newKeyAction("Filter", v.activateCmd, false)
	v.actions[KeyShiftR] = newKeyAction("Sort Resources", v.sortResourceCmd, true)
	v.actions[KeyShiftO] = newKeyAction("Sort Groups", v.sortGroupCmd, true)
}

func (v *aliasView) getTitle() string {
	return aliasTitle
}

func (v *aliasView) sortResourceCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.sortCol.index, v.sortCol.asc = 1, true
	v.refresh()
	return nil
}

func (v *aliasView) sortGroupCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.sortCol.index, v.sortCol.asc = 2, true
	v.refresh()
	return nil
}

func (v *aliasView) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.cmdBuff.empty() {
		v.cmdBuff.reset()
		return nil
	}

	return v.backCmd(evt)
}

func (v *aliasView) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	r, _ := v.GetSelection()
	if r != 0 {
		return v.runCmd(evt)
	}

	if v.cmdBuff.isActive() {
		return v.filterCmd(evt)
	}

	return evt
}

func (v *aliasView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cancel != nil {
		v.cancel()
	}

	if v.cmdBuff.isActive() {
		v.cmdBuff.reset()
	} else {
		v.app.inject(v.current)
	}

	return nil
}

func (v *aliasView) runCmd(evt *tcell.EventKey) *tcell.EventKey {
	r, _ := v.GetSelection()
	if r > 0 {
		v.app.gotoResource(strings.TrimSpace(v.GetCell(r, 0).Text), true)
	}

	return nil
}

func (v *aliasView) hints() hints {
	return v.actions.toHints()
}

func (v *aliasView) hydrate() resource.TableData {
	cmds := helpCmds(v.app.conn())

	data := resource.TableData{
		Header:    resource.Row{"NAME", "RESOURCE", "APIGROUP"},
		Rows:      make(resource.RowEvents, len(cmds)),
		Namespace: resource.NotNamespaced,
	}

	for k := range cmds {
		fields := resource.Row{
			resource.Pad(k, 30),
			resource.Pad(cmds[k].title, 30),
			resource.Pad(cmds[k].api, 30),
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
