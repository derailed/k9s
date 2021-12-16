package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell/v2"
)

// LogsExtender adds log actions to a given viewer.
type LogsExtender struct {
	ResourceViewer

	optionsFn LogOptionsFunc
}

// NewLogsExtender returns a new extender.
func NewLogsExtender(v ResourceViewer, f LogOptionsFunc) ResourceViewer {
	l := LogsExtender{
		ResourceViewer: v,
		optionsFn:      f,
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
	opts := l.buildLogOpts(path, "", prev)
	if l.optionsFn != nil {
		if opts, err = l.optionsFn(prev); err != nil {
			l.App().Flash().Err(err)
			return
		}
	}
	if err := l.App().inject(NewLog(l.GVR(), opts)); err != nil {
		l.App().Flash().Err(err)
	}
}

// buildLogOpts(path, co, prev, false, config.DefaultLoggerTailCount),.
func (l *LogsExtender) buildLogOpts(path, co string, prevLogs bool) *dao.LogOptions {
	cfg := l.App().Config.K9s.Logger
	opts := dao.LogOptions{
		Path:          path,
		Container:     co,
		Lines:         int64(cfg.TailCount),
		Previous:      prevLogs,
		ShowTimestamp: cfg.ShowTime,
	}
	if opts.Container == "" {
		opts.AllContainers = true
	}

	return &opts
}
