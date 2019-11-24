package view

import (
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

// LogsExtender adds log actions to a given viewer.
type LogsExtender struct {
	ResourceViewer

	containerFn ContainerFunc
}

// NewLogsExtender returns a new extender.
func NewLogsExtender(r ResourceViewer, f ContainerFunc) ResourceViewer {
	l := LogsExtender{
		ResourceViewer: r,
		containerFn:    f,
	}
	l.BindKeys()

	return &l
}

// BindKeys injects new menu actions.
func (l *LogsExtender) BindKeys() {
	l.Actions().Add(ui.KeyActions{
		ui.KeyL:      ui.NewKeyAction("Logs", l.logsCmd(false), true),
		ui.KeyShiftL: ui.NewKeyAction("Logs Previous", l.logsCmd(true), true),
	})
}

func (l *LogsExtender) logsCmd(prev bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		path := l.GetTable().GetSelectedItem()
		if path == "" {
			return nil
		}
		if l.GetTable().Path != "" {
			path = l.GetTable().Path
		}
		l.showLogs(path, prev)

		return nil
	}
}

func (l *LogsExtender) showLogs(path string, prev bool) {
	co := ""
	if l.containerFn != nil {
		co = l.containerFn()
	}
	log := NewLog(path, co, l.List(), prev)
	l.App().inject(log)
}
