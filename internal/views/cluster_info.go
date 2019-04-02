package views

import (
	"strings"

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

func newInfoView(app *appView) *clusterInfoView {
	return &clusterInfoView{
		app:     app,
		Table:   tview.NewTable(),
		cluster: resource.NewCluster(app.conn(), &log.Logger, k8s.NewMetricsServer(app.conn())),
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

func (*clusterInfoView) sectionCell(t string) *tview.TableCell {
	c := tview.NewTableCell(t + ":")
	c.SetAlign(tview.AlignLeft)

	return c
}

func (*clusterInfoView) infoCell(t string) *tview.TableCell {
	c := tview.NewTableCell(t)
	c.SetExpansion(2)
	c.SetTextColor(tcell.ColorOrange)

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

	nodes, err := v.cluster.FetchNodes()
	if err != nil {
		log.Warn().Msgf("ClusterInfo %s", err)
		return
	}
	mxNodes, err := v.cluster.FetchNodesMetrics()
	if err != nil {
		log.Warn().Msgf("ClusterInfo %s", err)
		return
	}

	mx := v.cluster.Metrics(nodes, mxNodes)
	c := v.GetCell(row, 1)
	cpu := toPerc(mx.PercCPU)
	stripTxt := strings.Replace(c.Text, plus(), "", 1)
	stripTxt = strings.Replace(stripTxt, minus(), "", 1)
	c.SetText(cpu + deltas(stripTxt, cpu))
	row++
	c = v.GetCell(row, 1)
	mem := toPerc(mx.PercMEM)
	c.SetText(mem + deltas(c.Text, mem))
}
