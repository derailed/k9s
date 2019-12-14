package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
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
		if l.GetTable().Path != "" {
			path = l.GetTable().Path
		}
		l.showLogs(path, prev)

		return nil
	}
}

func (l *LogsExtender) showLogs(path string, prev bool) {
	log.Debug().Msgf("SHOWING LOGS path %q", path)
	co := ""
	if l.containerFn != nil {
		log.Debug().Msgf("CUSTOM CO FUNC")
		co = l.containerFn()
	}
	if err := l.App().inject(NewLog(client.GVR(l.GVR()), path, co, prev)); err != nil {
		l.App().Flash().Err(err)
	}
}
