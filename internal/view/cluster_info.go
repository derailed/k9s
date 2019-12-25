package view

import (
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type clusterInfoView struct {
	*tview.Table

	app *App
	mxs *client.MetricsServer
}

func newClusterInfoView(app *App, mx *client.MetricsServer) *clusterInfoView {
	return &clusterInfoView{
		app:   app,
		Table: tview.NewTable(),
		mxs:   mx,
	}
}

func (v *clusterInfoView) init(version string) {
	cluster := model.NewCluster(v.app.Conn(), v.mxs)

	row := v.initInfo(cluster)
	row = v.initVersion(row, version, cluster)

	v.SetCell(row, 0, v.sectionCell("CPU"))
	v.SetCell(row, 1, v.infoCell(render.NAValue))
	row++
	v.SetCell(row, 0, v.sectionCell("MEM"))
	v.SetCell(row, 1, v.infoCell(render.NAValue))

	v.refresh()
}

func (v *clusterInfoView) initInfo(cluster *model.Cluster) int {
	var row int
	v.SetCell(row, 0, v.sectionCell("Context"))
	v.SetCell(row, 1, v.infoCell(cluster.ContextName()))
	row++

	v.SetCell(row, 0, v.sectionCell("Cluster"))
	v.SetCell(row, 1, v.infoCell(cluster.ClusterName()))
	row++

	v.SetCell(row, 0, v.sectionCell("User"))
	v.SetCell(row, 1, v.infoCell(cluster.UserName()))
	row++

	return row
}

func (v *clusterInfoView) initVersion(row int, version string, cluster *model.Cluster) int {
	v.SetCell(row, 0, v.sectionCell("K9s Rev"))
	v.SetCell(row, 1, v.infoCell(version))
	row++

	v.SetCell(row, 0, v.sectionCell("K8s Rev"))
	v.SetCell(row, 1, v.infoCell(cluster.Version()))
	row++

	return row
}

func (v *clusterInfoView) sectionCell(t string) *tview.TableCell {
	c := tview.NewTableCell(t + ":")
	c.SetAlign(tview.AlignLeft)
	var s tcell.Style
	c.SetStyle(s.Bold(true).Foreground(config.AsColor(v.app.Styles.K9s.Info.SectionColor)))
	c.SetBackgroundColor(v.app.Styles.BgColor())

	return c
}

func (v *clusterInfoView) infoCell(t string) *tview.TableCell {
	c := tview.NewTableCell(t)
	c.SetExpansion(2)
	c.SetTextColor(config.AsColor(v.app.Styles.K9s.Info.FgColor))
	c.SetBackgroundColor(v.app.Styles.BgColor())

	return c
}

func (v *clusterInfoView) refresh() {
	var (
		cluster = model.NewCluster(v.app.Conn(), v.mxs)
		row     int
	)
	v.GetCell(row, 1).SetText(cluster.ContextName())
	row++
	v.GetCell(row, 1).SetText(cluster.ClusterName())
	row++
	v.GetCell(row, 1).SetText(cluster.UserName())
	row += 2
	v.GetCell(row, 1).SetText(cluster.Version())
	row++

	c := v.GetCell(row, 1)
	c.SetText(render.NAValue)
	c = v.GetCell(row+1, 1)
	c.SetText(render.NAValue)

	v.refreshMetrics(cluster, row)
}

func fetchResources(app *App) (*v1.NodeList, *mv1beta1.NodeMetricsList, error) {
	nos, err := app.factory.Client().DialOrDie().CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	mx := client.NewMetricsServer(app.factory.Client())
	nmx, err := mx.FetchNodesMetrics()
	if err != nil {
		return nil, nil, err
	}

	return nos, nmx, nil
}

func (v *clusterInfoView) refreshMetrics(cluster *model.Cluster, row int) {
	nos, nmx, err := fetchResources(v.app)
	if err != nil {
		log.Warn().Msgf("NodeMetrics %#v", err)
		return
	}

	var cmx client.ClusterMetrics
	if err := cluster.Metrics(nos, nmx, &cmx); err != nil {
		log.Error().Err(err).Msgf("failed to retrieve cluster metrics")
	}
	c := v.GetCell(row, 1)
	cpu := render.AsPerc(cmx.PercCPU)
	if cpu == "0" {
		cpu = render.NAValue
	}
	c.SetText(cpu + "%" + ui.Deltas(strip(c.Text), cpu))
	row++

	c = v.GetCell(row, 1)
	mem := render.AsPerc(cmx.PercMEM)
	if mem == "0" {
		mem = render.NAValue
	}
	c.SetText(mem + "%" + ui.Deltas(strip(c.Text), mem))
}

func strip(s string) string {
	t := strings.Replace(s, ui.PlusSign, "", 1)
	t = strings.Replace(t, ui.MinusSign, "", 1)
	return t
}
