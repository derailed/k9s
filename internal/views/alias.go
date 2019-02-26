package views

import (
	"context"
	"fmt"
	"sort"
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
}

func newAliasView(app *appView) *aliasView {
	v := aliasView{tableView: newTableView(app, aliasTitle, nil)}
	{
		v.SetSelectedStyle(tcell.ColorWhite, tcell.ColorFuchsia, tcell.AttrNone)
		v.colorerFn = aliasColorer
		v.current = app.content.GetPrimitive("main").(igniter)
		v.sortFn = v.sorterFn
	}
	v.actions[tcell.KeyEnter] = newKeyAction("Search", v.aliasCmd)
	v.actions[tcell.KeyEscape] = newKeyAction("Reset", v.resetCmd)
	v.actions[KeySlash] = newKeyAction("Filter", v.activateCmd)
	return &v
}

func (v *aliasView) sorterFn(ss []string) {
	sort.Strings(ss)
}

// Init the view.
func (v *aliasView) init(context.Context, string) {
	v.update(v.hydrate())
	v.app.SetFocus(v)
	v.resetTitle()
}

func (v *aliasView) getTitle() string {
	return aliasTitle
}

func (v *aliasView) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.cmdBuff.empty() {
		v.cmdBuff.reset()
		v.refresh()
		return nil
	}
	return v.backCmd(evt)
}

func (v *aliasView) aliasCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cmdBuff.isActive() {
		return v.filterCmd(evt)
	}
	return v.runCmd(evt)
}

func (v *aliasView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.app.inject(v.current)
	return nil
}

func (v *aliasView) runCmd(evt *tcell.EventKey) *tcell.EventKey {
	r, _ := v.GetSelection()
	if r > 0 {
		v.app.command.run(strings.TrimSpace(v.GetCell(r, 0).Text))
	}
	return nil
}

func (v *aliasView) hints() hints {
	return v.actions.toHints()
}

func (v *aliasView) hydrate() resource.TableData {
	cmds := helpCmds()

	data := resource.TableData{
		Header:    resource.Row{"ALIAS", "RESOURCE", "APIGROUP"},
		Rows:      make(resource.RowEvents, len(cmds)),
		Namespace: "",
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
