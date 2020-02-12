package view

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
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

func (c *ClusterInfo) layout() {
	for row, v := range []string{"Context", "Cluster", "User", "K9s Rev", "K8s Rev", "CPU", "MEM"} {
		c.SetCell(row, 0, c.sectionCell(v))
		c.SetCell(row, 1, c.infoCell(render.NAValue))
	}
}

func (c *ClusterInfo) sectionCell(t string) *tview.TableCell {
	cell := tview.NewTableCell(t + ":")
	cell.SetAlign(tview.AlignLeft)
	// var style tcell.Style
	// style.Bold(true).
	// 	Background(tcell.ColorGreen).
	// 	Foreground(config.AsColor(c.styles.K9s.Info.SectionColor))
	// cell.SetStyle(style)
	// cell.SetBackgroundColor(c.app.Styles.BgColor())
	// cell.SetBackgroundColor(tcell.ColorDefault)
	cell.SetBackgroundColor(tcell.ColorGreen)

	return cell
}

func (c *ClusterInfo) infoCell(t string) *tview.TableCell {
	cell := tview.NewTableCell(t)
	cell.SetExpansion(2)
	cell.SetTextColor(config.AsColor(c.styles.K9s.Info.FgColor))
	cell.SetBackgroundColor(c.app.Styles.BgColor())

	return cell
}

// ClusterInfoUpdated notifies the cluster meta was updated.
func (c *ClusterInfo) ClusterInfoUpdated(data model.ClusterMeta) {
	c.app.QueueUpdateDraw(func() {
		var row int
		c.GetCell(row, 1).SetText(data.Context)
		row++
		c.GetCell(row, 1).SetText(data.Cluster)
		row++
		c.GetCell(row, 1).SetText(data.User)
		row++
		c.GetCell(row, 1).SetText(data.K9sVer)
		row++
		c.GetCell(row, 1).SetText(data.K8sVer)
		row++
		c.GetCell(row, 1).SetText(render.AsPerc(data.Cpu) + "%")
		row++
		c.GetCell(row, 1).SetText(render.AsPerc(data.Mem) + "%")

		c.updateStyle()
	})
}

// ClusterInfoChanged notifies the cluster meta was changed.
func (c *ClusterInfo) ClusterInfoChanged(prev, curr model.ClusterMeta) {
	c.app.QueueUpdateDraw(func() {
		var row int
		c.GetCell(row, 1).SetText(curr.Context)
		row++
		c.GetCell(row, 1).SetText(curr.Cluster)
		row++
		c.GetCell(row, 1).SetText(curr.User)
		row++
		c.GetCell(row, 1).SetText(curr.K9sVer)
		row++
		c.GetCell(row, 1).SetText(curr.K8sVer)
		row++
		c.GetCell(row, 1).SetText(ui.AsPercDelta(prev.Cpu, curr.Cpu))
		row++
		c.GetCell(row, 1).SetText(ui.AsPercDelta(prev.Mem, curr.Mem))

		c.updateStyle()
	})
}

func (c *ClusterInfo) updateStyle() {
	for row := 0; row < c.GetRowCount(); row++ {
		c.GetCell(row, 0).SetTextColor(config.AsColor(c.styles.K9s.Info.FgColor))
		c.GetCell(row, 0).SetBackgroundColor(c.styles.BgColor())
		var s tcell.Style
		c.GetCell(row, 1).SetStyle(s.Bold(true).Foreground(config.AsColor(c.styles.K9s.Info.SectionColor)))
	}
}
