package views

import (
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/watch"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type clusterInfoView struct {
	*tview.Table

	app *appView
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

func newClusterInfoView(app *appView, mx resource.MetricsServer) *clusterInfoView {
	return &clusterInfoView{
		app:   app,
		Table: tview.NewTable(),
		mxs:   mx,
	}
}

func (v *clusterInfoView) init() {
	cluster := resource.NewCluster(v.app.conn(), &log.Logger, v.mxs)

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

	v.SetCell(row, 0, v.sectionCell("K9s Rev"))
	v.SetCell(row, 1, v.infoCell(v.app.version))
	row++

	rev := cluster.Version()
	v.SetCell(row, 0, v.sectionCell("K8s Rev"))
	v.SetCell(row, 1, v.infoCell(rev))
	row++

	v.SetCell(row, 0, v.sectionCell("CPU"))
	v.SetCell(row, 1, v.infoCell(resource.NAValue))
	v.SetCell(row+1, 0, v.sectionCell("MEM"))
	v.SetCell(row+1, 1, v.infoCell(resource.NAValue))

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
	cluster := resource.NewCluster(v.app.conn(), &log.Logger, v.mxs)

	var row int
	v.GetCell(row, 1).SetText(cluster.ContextName())
	row++
	v.GetCell(row, 1).SetText(cluster.ClusterName())
	row++
	v.GetCell(row, 1).SetText(cluster.UserName())
	row += 2
	v.GetCell(row, 1).SetText(cluster.Version())
	row++

	nos, err := v.app.informer.List(watch.NodeIndex, "", metav1.ListOptions{})
	if err != nil {
		log.Warn().Err(err).Msg("List nodes")
		return
	}
	nmx, err := v.app.informer.List(watch.NodeMXIndex, "", metav1.ListOptions{})
	if err != nil {
		log.Warn().Err(err).Msg("List node metrics")
		return
	}
	var cmx k8s.ClusterMetrics
	cluster.Metrics(nos, nmx, &cmx)
	c := v.GetCell(row, 1)
	cpu := resource.AsPerc(cmx.PercCPU)
	c.SetText(cpu + deltas(strip(c.Text), cpu))
	row++

	c = v.GetCell(row, 1)
	mem := resource.AsPerc(cmx.PercMEM)
	c.SetText(mem + deltas(strip(c.Text), mem))
}

func strip(s string) string {
	t := strings.Replace(s, plus(), "", 1)
	t = strings.Replace(t, minus(), "", 1)
	return t
}
