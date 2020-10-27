package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

// LogsExtender adds log actions to a given viewer.
type LogsExtender struct {
	ResourceViewer

	containerFn ContainerFunc
}

// NewLogsExtender returns a new extender.
func NewLogsExtender(v ResourceViewer, f ContainerFunc) ResourceViewer {
	l := LogsExtender{
		ResourceViewer: v,
		containerFn:    f,
	}
	l.AddBindKeysFn(l.bindKeys)

	return &l
}

// BindKeys injects new menu actions.
func (l *LogsExtender) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyL: ui.NewKeyAction("Logs", l.logsCmd(false), true),
		ui.KeyP: ui.NewKeyAction("Logs Previous", l.logsCmd(true), true),
	})
}

func (l *LogsExtender) logsCmd(prev bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		path := l.GetTable().GetSelectedItem()
		if path == "" {
			return nil
		}
		if !isResourcePath(path) {
			path = l.GetTable().Path
		}
		l.showLogs(path, prev)

		return nil
	}
}

func isResourcePath(p string) bool {
	ns, n := client.Namespaced(p)
	return ns != "" && n != ""
}

func (l *LogsExtender) showLogs(path string, prev bool) {
	ns, _ := client.Namespaced(path)
	_, err := l.App().factory.CanForResource(ns, "v1/pods", client.MonitorAccess)
	if err != nil {
		l.App().Flash().Err(err)
		return
	}

	co := ""
	if l.containerFn != nil {
		co = l.containerFn()
	}
	if err := l.App().inject(NewLog(l.GVR(), path, co, prev)); err != nil {
		l.App().Flash().Err(err)
	}
}
