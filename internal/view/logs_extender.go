package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/watch"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
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
	l.bindKeys(l.Actions())

	return &l
}

// BindKeys injects new menu actions.
func (l *LogsExtender) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
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
	log.Debug().Msgf("SHOWING LOGS path %q", path)
	// Need to load and wait for pods
	ns, _ := render.Namespaced(path)
	_, err := l.App().factory.CanForResource(ns, "v1/pods", watch.ReadVerbs)
	if err != nil {
		l.App().Flash().Err(err)
		return
	}

	co := ""
	if l.containerFn != nil {
		co = l.containerFn()
	}
	if err := l.App().inject(NewLog(client.NewGVR(l.GVR()), path, co, prev)); err != nil {
		l.App().Flash().Err(err)
	}
}
