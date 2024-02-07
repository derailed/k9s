// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// StatusIndicator represents a status indicator when main header is collapsed.
type StatusIndicator struct {
	*tview.TextView

	app       *App
	styles    *config.Styles
	permanent string
	cancel    context.CancelFunc
}

// NewStatusIndicator returns a new status indicator.
func NewStatusIndicator(app *App, styles *config.Styles) *StatusIndicator {
	s := StatusIndicator{
		TextView: tview.NewTextView(),
		app:      app,
		styles:   styles,
	}
	s.CorrectStatusIndConfig()
	s.SetTextAlign(tview.AlignCenter)
	s.SetTextColor(tcell.ColorWhite)
	s.SetBackgroundColor(styles.BgColor())
	s.SetDynamicColors(true)
	styles.AddListener(&s)

	return &s
}

// StylesChanged notifies the skins changed.
func (s *StatusIndicator) StylesChanged(styles *config.Styles) {
	s.styles = styles
	s.SetBackgroundColor(styles.BgColor())
	s.SetTextColor(styles.FgColor())
}

const defaultStatusIndicatorFmt = "[orange::b]K9s [aqua::]%s [white::]%s:%s:%s [lawngreen::]%s[white::]::[darkturquoise::]%s"

// Checks if status indicator config is correct if not loads default.
func (s *StatusIndicator) CorrectStatusIndConfig() {
	statIndConf := &s.app.Config.K9s.UI.StatusIndicator
	if statIndConf.Format == "" || len(statIndConf.Fields) != strings.Count(statIndConf.Format, "%s") {
		statIndConf.Format = defaultStatusIndicatorFmt
		statIndConf.Fields = []string{"K9SVER", "CONTEXT", "CLUSTER", "K8SVER", "CPU", "MEMORY"}
	}
}

// ClusterInfoUpdated notifies the cluster meta was updated.
func (s *StatusIndicator) ClusterInfoUpdated(data model.ClusterMeta) {
	s.app.QueueUpdateDraw(func() {
		s.SetPermanent(
			s.BuildStatusIndicatorText(nil, &data),
		)
	})
}

// Builds the text to be put into the status indicator based on the config.
func (s *StatusIndicator) BuildStatusIndicatorText(prev, cur *model.ClusterMeta) (statusStr string) {
	var (
		statIndConf config.StatusIndicatorConf
		cpuPerc     string
		memPerc     string
		orderedData []any
	)

	if prev != nil {
		cpuPerc = AsPercDelta(prev.Cpu, cur.Cpu)
		memPerc = AsPercDelta(prev.Cpu, cur.Mem)
	} else {
		cpuPerc = render.PrintPerc(cur.Cpu)
		memPerc = render.PrintPerc(cur.Mem)
	}

	statIndConf = s.app.Config.K9s.UI.StatusIndicator
	for _, field := range statIndConf.Fields {
		switch field {
		case "K9SVER":
			orderedData = append(orderedData, cur.K9sVer)
		case "CONTEXT":
			orderedData = append(orderedData, cur.Context)
		case "CLUSTER":
			orderedData = append(orderedData, cur.Cluster)
		case "USER":
			orderedData = append(orderedData, cur.User)
		case "K8SVER":
			orderedData = append(orderedData, cur.K8sVer)
		case "CPU":
			orderedData = append(orderedData, cpuPerc)
		case "MEMORY":
			orderedData = append(orderedData, memPerc)
		}
	}
	return fmt.Sprintf(statIndConf.Format, orderedData...)
}

// ClusterInfoChanged notifies the cluster meta was changed.
func (s *StatusIndicator) ClusterInfoChanged(prev, cur model.ClusterMeta) {
	if !s.app.IsRunning() {
		return
	}
	s.app.QueueUpdateDraw(func() {
		s.SetPermanent(s.BuildStatusIndicatorText(&prev, &cur))
	})
}

// SetPermanent sets permanent title to be reset to after updates.
func (s *StatusIndicator) SetPermanent(info string) {
	s.permanent = info
	s.SetText(info)
}

// Reset clears out the logo view and resets colors.
func (s *StatusIndicator) Reset() {
	s.Clear()
	s.SetPermanent(s.permanent)
}

// Err displays a log error state.
func (s *StatusIndicator) Err(msg string) {
	s.update(msg, "orangered")
}

// Warn displays a log warning state.
func (s *StatusIndicator) Warn(msg string) {
	s.update(msg, "mediumvioletred")
}

// Info displays a log info state.
func (s *StatusIndicator) Info(msg string) {
	s.update(msg, "lawngreen")
}

func (s *StatusIndicator) update(msg, c string) {
	s.setText(fmt.Sprintf("[%s::b] <%s> ", c, msg))
}

func (s *StatusIndicator) setText(msg string) {
	if s.cancel != nil {
		s.cancel()
	}
	s.SetText(msg)

	var ctx context.Context
	ctx, s.cancel = context.WithCancel(context.Background())
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			s.app.QueueUpdateDraw(func() {
				s.Reset()
			})
		}
	}(ctx)
}

// Helpers...

// AsPercDelta represents a percentage with a delta indicator.
func AsPercDelta(ov, nv int) string {
	prev, cur := render.IntToStr(ov), render.IntToStr(nv)
	return cur + "%" + Deltas(prev, cur)
}
