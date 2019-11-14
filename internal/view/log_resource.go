package view

import (
	"context"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

// ContainerFn returns the active container name.
type containerFn func() string

// LogResource represents a loggable resource view.
type LogResource struct {
	*Resource

	containerFn containerFn
	logs        *Logs
}

func NewLogResource(title, gvr string, list resource.List) *LogResource {
	l := LogResource{
		Resource: NewResource(title, gvr, list),
	}
	l.logs = NewLogs(list.GetName(), &l)

	return &l
}

func (l *LogResource) Init(ctx context.Context) {
	l.Resource.Init(ctx)
	l.logs.Init(ctx)
}

func (l *LogResource) extraActions(aa ui.KeyActions) {
	aa[ui.KeyL] = ui.NewKeyAction("Logs", l.logsCmd, true)
	aa[ui.KeyShiftL] = ui.NewKeyAction("Logs Previous", l.prevLogsCmd, true)
}

func (l *LogResource) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := l.masterPage()
		t.SetSortCol(t.NameColIndex()+col, 0, asc)
		t.Refresh()

		return nil
	}
}

// Protocol...

func (l *LogResource) getList() resource.List {
	return l.list
}

func (l *LogResource) getSelection() string {
	if l.path != nil {
		return *l.path
	}
	return l.masterPage().GetSelectedItem()
}

func (l *LogResource) prevLogsCmd(evt *tcell.EventKey) *tcell.EventKey {
	l.showLogs(true)
	return nil
}

func (l *LogResource) logsCmd(evt *tcell.EventKey) *tcell.EventKey {
	l.showLogs(false)
	return nil
}

func (l *LogResource) showLogs(prev bool) {
	if !l.masterPage().RowSelected() {
		return
	}

	co := ""
	if l.containerFn != nil {
		co = l.containerFn()
	}
	l.logs.reload(co, l, prev)
	l.Push(l.logs)
}

func (l *LogResource) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if err := l.app.Config.SetActiveNamespace(l.list.GetNamespace()); err != nil {
		log.Error().Err(err).Msg("Config NS set failed!")
	}
	l.app.inject(l)

	return nil
}
