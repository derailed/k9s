package views

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
)

type (
	containerFn func() string

	logResourceView struct {
		*resourceView

		containerFn containerFn
	}
)

func newLogResourceView(ns string, app *appView, list resource.List) *logResourceView {
	v := logResourceView{
		resourceView: newResourceView(ns, app, list).(*resourceView),
	}
	v.AddPage("logs", newLogsView(list.GetName(), app, &v), true, false)

	return &v
}

func (v *logResourceView) extraActions(aa keyActions) {
	aa[KeyL] = newKeyAction("Logs", v.logsCmd, true)
	aa[KeyShiftL] = newKeyAction("Logs Previous", v.prevLogsCmd, true)
}

func (v *logResourceView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.getTV()
		t.sortCol.index, t.sortCol.asc = t.nameColIndex()+col, asc
		t.refresh()

		return nil
	}
}

// Protocol...

func (v *logResourceView) getList() resource.List {
	return v.list
}

func (v *logResourceView) getSelection() string {
	if v.path != nil {
		return *v.path
	}
	return v.selectedItem
}

func (v *logResourceView) prevLogsCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.showLogs(true)

	return nil
}

func (v *logResourceView) logsCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.showLogs(false)

	return nil
}

func (v *logResourceView) showLogs(prev bool) {
	if !v.rowSelected() {
		return
	}

	l := v.GetPrimitive("logs").(*logsView)
	co := ""
	if v.containerFn != nil {
		co = v.containerFn()
	}
	l.reload(co, v, v.list.GetName(), prev)
	v.switchPage("logs")
}

func (v *logResourceView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	// Reset namespace to what it was
	v.app.config.SetActiveNamespace(v.list.GetNamespace())
	v.app.inject(v)

	return nil
}
