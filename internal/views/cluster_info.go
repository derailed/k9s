package views

import (
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

type clusterInfoView struct {
	*tview.Table

	app     *appView
	cluster *resource.Cluster
}

// ClusterInfo tracks Kubernetes cluster and K9s information.
type ClusterInfo interface {
	ContextName() string
	ClusterName() string
	UserName() string
	K9sVersion() string
	K8sVersion() string
	CurrentCPU() float64
	CurrentMEM() float64
}

func newClusterInfoView(app *appView, mx resource.MetricsServer) *clusterInfoView {
	return &clusterInfoView{
		app:     app,
		Table:   tview.NewTable(),
		cluster: resource.NewCluster(app.conn(), &log.Logger, mx),
	}
}

func (v *clusterInfoView) init() {
	var row int
	v.SetCell(row, 0, v.sectionCell("Context"))
	v.SetCell(row, 1, v.infoCell(v.cluster.ContextName()))
	row++

	v.SetCell(row, 0, v.sectionCell("Cluster"))
	v.SetCell(row, 1, v.infoCell(v.cluster.ClusterName()))
	row++

	v.SetCell(row, 0, v.sectionCell("User"))
	v.SetCell(row, 1, v.infoCell(v.cluster.UserName()))
	row++

	v.SetCell(row, 0, v.sectionCell("K9s Rev"))
	v.SetCell(row, 1, v.infoCell(v.app.version))
	row++

	rev := v.cluster.Version()
	v.SetCell(row, 0, v.sectionCell("K8s Rev"))
	v.SetCell(row, 1, v.infoCell(rev))
	row++

	v.SetCell(row, 0, v.sectionCell("CPU"))
	v.SetCell(row, 1, v.infoCell("n/a"))
	v.SetCell(row+1, 0, v.sectionCell("MEM"))
	v.SetCell(row+1, 1, v.infoCell("n/a"))

	v.refresh()
}

func (v *clusterInfoView) sectionCell(t string) *tview.TableCell {
	c := tview.NewTableCell(t + ":")
	c.SetAlign(tview.AlignLeft)
	var s tcell.Style
	c.SetStyle(s.Bold(true).Foreground(config.AsColor(v.app.styles.Style.Info.SectionColor)))
	c.SetBackgroundColor(v.app.styles.BgColor())

	return c
}

func (v *clusterInfoView) infoCell(t string) *tview.TableCell {
	c := tview.NewTableCell(t)
	c.SetExpansion(2)
	c.SetTextColor(config.AsColor(v.app.styles.Style.Info.FgColor))
	c.SetBackgroundColor(v.app.styles.BgColor())

	return c
}

func (v *clusterInfoView) refresh() {
	var row int
	v.GetCell(row, 1).SetText(v.cluster.ContextName())
	row++
	v.GetCell(row, 1).SetText(v.cluster.ClusterName())
	row++
	v.GetCell(row, 1).SetText(v.cluster.UserName())
	row += 2
	v.GetCell(row, 1).SetText(v.cluster.Version())
	row++

	nodes, err := v.cluster.GetNodes()
	if err != nil {
		log.Warn().Err(err).Msg("List nodes")
		return
	}
	mxNodes, err := v.cluster.FetchNodesMetrics()
	if err != nil {
		log.Warn().Err(err).Msg("List node metrics")
		return
	}

	var mx k8s.ClusterMetrics
	v.cluster.Metrics(nodes, mxNodes, &mx)
	c := v.GetCell(row, 1)
	cpu := resource.AsPerc(mx.PercCPU)
	c.SetText(cpu + deltas(strip(c.Text), cpu))
	row++

	c = v.GetCell(row, 1)
	mem := resource.AsPerc(mx.PercMEM)
	c.SetText(mem + deltas(strip(c.Text), mem))
}

func strip(s string) string {
	t := strings.Replace(s, plus(), "", 1)
	t = strings.Replace(t, minus(), "", 1)
	return t
}
