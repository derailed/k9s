// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
func (l *LogsExtender) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
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
	_, err := l.App().factory.CanForResource(ns, "v1/pods", client.ListAccess)
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
	if err := l.App().inject(NewLog(l.GVR(), opts), false); err != nil {
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

func podLogOptions(app *App, fqn string, prev bool, m metav1.ObjectMeta, spec v1.PodSpec) *dao.LogOptions {
	var (
		cc   = fetchContainers(m, spec, true)
		cfg  = app.Config.K9s.Logger
		opts = dao.LogOptions{
			Path:            fqn,
			Lines:           int64(cfg.TailCount),
			SinceSeconds:    cfg.SinceSeconds,
			SingleContainer: len(cc) == 1,
			ShowTimestamp:   cfg.ShowTime,
			Previous:        prev,
		}
	)
	if c, ok := dao.GetDefaultContainer(m, spec); ok {
		opts.Container, opts.DefaultContainer = c, c
	} else if len(cc) == 1 {
		opts.Container = cc[0]
	} else {
		opts.AllContainers = true
	}

	return &opts
}
