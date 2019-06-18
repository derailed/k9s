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
	aliasTitleFmt = " [aqua::b]%s([fuchsia::b]%d[fuchsia::-])[aqua::-] "
)

type aliasView struct {
	*tableView

	current igniter
	cancel  context.CancelFunc
}

func newAliasView(app *appView, current igniter) *aliasView {
	v := aliasView{tableView: newTableView(app, aliasTitle)}
	{
		v.SetBorderFocusColor(tcell.ColorFuchsia)
		v.SetSelectedStyle(tcell.ColorWhite, tcell.ColorFuchsia, tcell.AttrNone)
		v.colorerFn = aliasColorer
		v.current = current
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
	v.app.setHints(v.hints())

}

func (v *aliasView) registerActions() {
	delete(v.actions, KeyShiftA)
	v.actions[tcell.KeyEnter] = newKeyAction("Goto", v.gotoCmd, true)
	v.actions[tcell.KeyEscape] = newKeyAction("Reset", v.resetCmd, false)
	v.actions[KeySlash] = newKeyAction("Filter", v.activateCmd, false)
	v.actions[KeyShiftR] = newKeyAction("Sort Resources", v.sortColCmd(1), true)
	v.actions[KeyShiftO] = newKeyAction("Sort Groups", v.sortColCmd(2), true)
}

func (v *aliasView) getTitle() string {
	return aliasTitle
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
	cmds := make(map[string]resCmd, 40)
	aliasCmds(v.app.conn(), cmds)

	data := resource.TableData{
		Header:    resource.Row{"NAME", "RESOURCE", "APIGROUP"},
		Rows:      make(resource.RowEvents, len(cmds)),
		Namespace: resource.NotNamespaced,
	}

	for k := range cmds {
		fields := resource.Row{
			pad(k, 30),
			pad(cmds[k].title, 30),
			pad(cmds[k].api, 30),
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
