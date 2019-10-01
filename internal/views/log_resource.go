package views

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

type (
	containerFn func() string

	logResourceView struct {
		*resourceView

		containerFn containerFn
	}
)

func newLogResourceView(title, gvr string, app *appView, list resource.List) *logResourceView {
	v := logResourceView{
		resourceView: newResourceView(title, gvr, app, list).(*resourceView),
	}
	v.AddPage("logs", newLogsView(list.GetName(), app, &v), true, false)

	return &v
}

func (v *logResourceView) extraActions(aa ui.KeyActions) {
	aa[ui.KeyL] = ui.NewKeyAction("Logs", v.logsCmd, true)
	aa[ui.KeyShiftL] = ui.NewKeyAction("Logs Previous", v.prevLogsCmd, true)
}

func (v *logResourceView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.masterPage()
		t.SetSortCol(t.NameColIndex()+col, 0, asc)
		t.Refresh()

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
	return v.masterPage().GetSelectedItem()
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
	if !v.masterPage().RowSelected() {
		return
	}

	l := v.GetPrimitive("logs").(*logsView)
	co := ""
	if v.containerFn != nil {
		co = v.containerFn()
	}
	l.reload(co, v, prev)
	v.switchPage("logs")
}

func (v *logResourceView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	// Reset namespace to what it was
	v.app.Config.SetActiveNamespace(v.list.GetNamespace())
	v.app.inject(v)

	return nil
}
