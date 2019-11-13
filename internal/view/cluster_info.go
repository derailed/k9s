package view

import (
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/watch"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type clusterInfoView struct {
	*tview.Table

	app *App
	mxs resource.MetricsServer
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

func newClusterInfoView(app *App, mx resource.MetricsServer) *clusterInfoView {
	return &clusterInfoView{
		app:   app,
		Table: tview.NewTable(),
		mxs:   mx,
	}
}

func (v *clusterInfoView) init(version string) {
	cluster := resource.NewCluster(v.app.Conn(), &log.Logger, v.mxs)

	row := v.initInfo(version, cluster)
	row = v.initVersion(row, version, cluster)

	v.SetCell(row, 0, v.sectionCell("CPU"))
	v.SetCell(row, 1, v.infoCell(resource.NAValue))
	row++
	v.SetCell(row, 0, v.sectionCell("MEM"))
	v.SetCell(row, 1, v.infoCell(resource.NAValue))

	v.refresh()
}

func (v *clusterInfoView) initInfo(version string, cluster *resource.Cluster) int {
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

func (v *clusterInfoView) initVersion(row int, version string, cluster *resource.Cluster) int {
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
		cluster = resource.NewCluster(v.app.Conn(), &log.Logger, v.mxs)
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
	c.SetText(resource.NAValue)
	c = v.GetCell(row+1, 1)
	c.SetText(resource.NAValue)

	v.refreshMetrics(cluster, row)
}

func fetchResources(app *App) (k8s.Collection, k8s.Collection, error) {
	nos, err := app.informer.List(watch.NodeIndex, "", metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	nmx, err := app.informer.List(watch.NodeMXIndex, "", metav1.ListOptions{})
	if err != nil {
		return nil, nil, err
	}

	return nos, nmx, nil
}

func (v *clusterInfoView) refreshMetrics(cluster *resource.Cluster, row int) {
	nos, nmx, err := fetchResources(v.app)
	if err != nil {
		log.Warn().Msgf("NodeMetrics %#v", err)
		return
	}

	var cmx k8s.ClusterMetrics
	cluster.Metrics(nos, nmx, &cmx)
	c := v.GetCell(row, 1)
	cpu := resource.AsPerc(cmx.PercCPU)
	if cpu == "0" {
		cpu = resource.NAValue
	}
	c.SetText(cpu + "%" + ui.Deltas(strip(c.Text), cpu))
	row++

	c = v.GetCell(row, 1)
	mem := resource.AsPerc(cmx.PercMEM)
	if mem == "0" {
		mem = resource.NAValue
	}
	c.SetText(mem + "%" + ui.Deltas(strip(c.Text), mem))
}

func strip(s string) string {
	t := strings.Replace(s, ui.PlusSign, "", 1)
	t = strings.Replace(t, ui.MinusSign, "", 1)
	return t
}
