package view

import (
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// ClusterInfo represents a cluster info view.
type ClusterInfo struct {
	*tview.Table

	app    *App
	mxs    *client.MetricsServer
	styles *config.Styles
}

// NewClusterInfo returns a new cluster info view.
func NewClusterInfo(app *App, mx *client.MetricsServer) *ClusterInfo {
	return &ClusterInfo{
		app:    app,
		Table:  tview.NewTable(),
		mxs:    mx,
		styles: app.Styles,
	}
}

// Init initializes the view.
func (c *ClusterInfo) Init(version string) {
	cluster := model.NewCluster(c.app.Conn(), c.mxs)

	c.app.Styles.AddListener(c)

	row := c.initInfo(cluster)
	row = c.initVersion(row, version, cluster)

	c.SetCell(row, 0, c.sectionCell("CPU"))
	c.SetCell(row, 1, c.infoCell(render.NAValue))
	row++
	c.SetCell(row, 0, c.sectionCell("MEM"))
	c.SetCell(row, 1, c.infoCell(render.NAValue))

	c.refresh()
}

// StylesChanged notifies skin changed.
func (c *ClusterInfo) StylesChanged(s *config.Styles) {
	c.styles = s
	c.SetBackgroundColor(s.BgColor())
	c.refresh()
}

func (c *ClusterInfo) initInfo(cluster *model.Cluster) int {
	var row int
	c.SetCell(row, 0, c.sectionCell("Context"))
	c.SetCell(row, 1, c.infoCell(cluster.ContextName()))
	row++

	c.SetCell(row, 0, c.sectionCell("Cluster"))
	c.SetCell(row, 1, c.infoCell(cluster.ClusterName()))
	row++

	c.SetCell(row, 0, c.sectionCell("User"))
	c.SetCell(row, 1, c.infoCell(cluster.UserName()))
	row++

	return row
}

func (c *ClusterInfo) initVersion(row int, version string, cluster *model.Cluster) int {
	c.SetCell(row, 0, c.sectionCell("K9s Rev"))
	c.SetCell(row, 1, c.infoCell(version))
	row++

	c.SetCell(row, 0, c.sectionCell("K8s Rev"))
	c.SetCell(row, 1, c.infoCell(cluster.Version()))
	row++

	return row
}

func (c *ClusterInfo) sectionCell(t string) *tview.TableCell {
	cell := tview.NewTableCell(t + ":")
	cell.SetAlign(tview.AlignLeft)
	var s tcell.Style
	cell.SetStyle(s.Bold(true).Foreground(config.AsColor(c.styles.K9s.Info.SectionColor)))
	cell.SetBackgroundColor(c.app.Styles.BgColor())

	return cell
}

func (c *ClusterInfo) infoCell(t string) *tview.TableCell {
	cell := tview.NewTableCell(t)
	cell.SetExpansion(2)
	cell.SetTextColor(config.AsColor(c.styles.K9s.Info.FgColor))
	cell.SetBackgroundColor(c.app.Styles.BgColor())

	return cell
}

func (c *ClusterInfo) refresh() {
	var (
		cluster = model.NewCluster(c.app.Conn(), c.mxs)
		row     int
	)

	c.GetCell(row, 1).SetText(cluster.ContextName())
	row++
	c.GetCell(row, 1).SetText(cluster.ClusterName())
	row++
	c.GetCell(row, 1).SetText(cluster.UserName())
	row += 2
	c.GetCell(row, 1).SetText(cluster.Version())
	row++

	cell := c.GetCell(row, 1)
	cell.SetText(render.NAValue)
	cell = c.GetCell(row+1, 1)
	cell.SetText(render.NAValue)

	if c.app.Conn().HasMetrics() {
		c.refreshMetrics(cluster, row)
	}
	c.updateStyle()
}

func (c *ClusterInfo) updateStyle() {
	for row := 0; row < c.GetRowCount(); row++ {
		c.GetCell(row, 0).SetTextColor(config.AsColor(c.styles.K9s.Info.FgColor))
		c.GetCell(row, 0).SetBackgroundColor(c.styles.BgColor())
		var s tcell.Style
		c.GetCell(row, 1).SetStyle(s.Bold(true).Foreground(config.AsColor(c.styles.K9s.Info.SectionColor)))
	}
}

func fetchResources(app *App) (*v1.NodeList, *mv1beta1.NodeMetricsList, error) {
	nn, err := dao.FetchNodes(app.factory, "")
	if err != nil {
		return nil, nil, err
	}

	mx := client.NewMetricsServer(app.factory.Client())
	nmx, err := mx.FetchNodesMetrics()
	if err != nil {
		return nil, nil, err
	}

	return nn, nmx, nil
}

func (c *ClusterInfo) refreshMetrics(cluster *model.Cluster, row int) {
	nos, nmx, err := fetchResources(c.app)
	if err != nil {
		log.Warn().Msgf("NodeMetrics %#v", err)
		return
	}

	var cmx client.ClusterMetrics
	if err := cluster.Metrics(nos, nmx, &cmx); err != nil {
		log.Error().Err(err).Msgf("failed to retrieve cluster metrics")
	}
	cell := c.GetCell(row, 1)
	cpu := render.AsPerc(cmx.PercCPU)
	if cpu == "0" {
		cpu = render.NAValue
	}
	cell.SetText(cpu + "%" + ui.Deltas(strip(cell.Text), cpu))
	row++

	cell = c.GetCell(row, 1)
	mem := render.AsPerc(cmx.PercMEM)
	if mem == "0" {
		mem = render.NAValue
	}
	cell.SetText(mem + "%" + ui.Deltas(strip(cell.Text), mem))
}

func strip(s string) string {
	t := strings.Replace(s, ui.PlusSign, "", 1)
	t = strings.Replace(t, ui.MinusSign, "", 1)
	return t
}
