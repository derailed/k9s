// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
)

var _ model.ClusterInfoListener = (*ClusterInfo)(nil)

// ClusterInfo represents a cluster info view.
type ClusterInfo struct {
	*tview.Table

	app    *App
	styles *config.Styles
}

// NewClusterInfo returns a new cluster info view.
func NewClusterInfo(app *App) *ClusterInfo {
	return &ClusterInfo{
		Table:  tview.NewTable(),
		app:    app,
		styles: app.Styles,
	}
}

// Init initializes the view.
func (c *ClusterInfo) Init() {
	c.SetBorderPadding(0, 0, 1, 0)
	c.app.Styles.AddListener(c)
	c.layout()
	c.StylesChanged(c.app.Styles)
}

// StylesChanged notifies skin changed.
func (c *ClusterInfo) StylesChanged(s *config.Styles) {
	c.styles = s
	c.SetBackgroundColor(s.BgColor())
	c.updateStyle()
}

func (c *ClusterInfo) hasMetrics() bool {
	mx := c.app.Conn().HasMetrics()
	if mx {
		auth, err := c.app.Conn().CanI("", "metrics.k8s.io/v1beta1/nodes", "", client.ListAccess)
		if err != nil {
			log.Warn().Err(err).Msgf("No nodes metrics access")
		}
		mx = auth
	}

	return mx
}

func (c *ClusterInfo) layout() {
	for row, section := range []string{"Context", "Cluster", "User", "K9s Rev", "K8s Rev", "CPU", "MEM"} {
		c.SetCell(row, 0, c.sectionCell(section))
		c.SetCell(row, 1, c.infoCell(render.NAValue))
	}
}

func (c *ClusterInfo) sectionCell(t string) *tview.TableCell {
	cell := tview.NewTableCell(t + ":")
	cell.SetAlign(tview.AlignLeft)
	cell.SetBackgroundColor(c.app.Styles.BgColor())

	return cell
}

func (c *ClusterInfo) infoCell(t string) *tview.TableCell {
	cell := tview.NewTableCell(t)
	cell.SetExpansion(2)
	cell.SetTextColor(c.styles.K9s.Info.FgColor.Color())
	cell.SetBackgroundColor(c.app.Styles.BgColor())

	return cell
}

func (c *ClusterInfo) setCell(row int, s string) int {
	if s == "" {
		s = render.NAValue
	}
	c.GetCell(row, 1).SetText(s)
	return row + 1
}

// ClusterInfoUpdated notifies the cluster meta was updated.
func (c *ClusterInfo) ClusterInfoUpdated(data model.ClusterMeta) {
	c.ClusterInfoChanged(data, data)
}

func (c *ClusterInfo) warnCell(s string, w bool) string {
	if w {
		return fmt.Sprintf("[orangered::b]%s", s)
	}

	return s
}

// ClusterInfoChanged notifies the cluster meta was changed.
func (c *ClusterInfo) ClusterInfoChanged(prev, curr model.ClusterMeta) {
	c.app.QueueUpdateDraw(func() {
		c.Clear()
		c.layout()
		row := c.setCell(0, curr.Context)
		row = c.setCell(row, curr.Cluster)
		row = c.setCell(row, curr.User)
		if curr.K9sLatest != "" {
			row = c.setCell(row, fmt.Sprintf("%s ⚡️[cadetblue::b]%s", curr.K9sVer, curr.K9sLatest))
		} else {
			row = c.setCell(row, curr.K9sVer)
		}
		row = c.setCell(row, curr.K8sVer)
		if c.hasMetrics() {
			row = c.setCell(row, ui.AsPercDelta(prev.Cpu, curr.Cpu))
			_ = c.setCell(row, ui.AsPercDelta(prev.Mem, curr.Mem))
			c.setDefCon(curr.Cpu, curr.Mem)
		} else {
			row = c.setCell(row, c.warnCell(render.NAValue, true))
			_ = c.setCell(row, c.warnCell(render.NAValue, true))
		}
		c.updateStyle()
	})
}

const defconFmt = "%s %s level!"

func (c *ClusterInfo) setDefCon(cpu, mem int) {
	var set bool
	l := c.app.Config.K9s.Thresholds.LevelFor("cpu", cpu)
	if l > config.SeverityLow {
		c.app.Status(flashLevel(l), fmt.Sprintf(defconFmt, flashMessage(l), "CPU"))
		set = true
	}
	l = c.app.Config.K9s.Thresholds.LevelFor("memory", mem)
	if l > config.SeverityLow {
		c.app.Status(flashLevel(l), fmt.Sprintf(defconFmt, flashMessage(l), "Memory"))
		set = true
	}
	if !set && !c.app.IsBenchmarking() {
		c.app.ClearStatus(true)
	}
}

func (c *ClusterInfo) updateStyle() {
	for row := 0; row < c.GetRowCount(); row++ {
		c.GetCell(row, 0).SetTextColor(c.styles.K9s.Info.FgColor.Color())
		c.GetCell(row, 0).SetBackgroundColor(c.styles.BgColor())
		var s tcell.Style
		s = s.Bold(true)
		s = s.Foreground(c.styles.K9s.Info.SectionColor.Color())
		s = s.Background(c.styles.BgColor())
		c.GetCell(row, 1).SetStyle(s)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func flashLevel(l config.SeverityLevel) model.FlashLevel {
	// nolint:exhaustive
	switch l {
	case config.SeverityHigh:
		return model.FlashErr
	case config.SeverityMedium:
		return model.FlashWarn
	default:
		return model.FlashInfo
	}
}

func flashMessage(l config.SeverityLevel) string {
	// nolint:exhaustive
	switch l {
	case config.SeverityHigh:
		return "Critical"
	case config.SeverityMedium:
		return "Warning"
	default:
		return "OK"
	}
}
