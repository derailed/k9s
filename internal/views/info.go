package views

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

type infoView struct {
	*tview.Table

	app *appView
}

func newInfoView(app *appView) *infoView {
	return &infoView{app: app, Table: tview.NewTable()}
}

func (v *infoView) init() {
	cluster := resource.NewCluster()

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

	v.SetCell(row, 0, v.sectionCell("K9s Version"))
	v.SetCell(row, 1, v.infoCell(v.app.version))
	row++

	rev := cluster.Version()
	v.SetCell(row, 0, v.sectionCell("K8s Version"))
	v.SetCell(row, 1, v.infoCell(rev))
	row++

	v.SetCell(row, 0, v.sectionCell("CPU"))
	v.SetCell(row, 1, v.infoCell("n/a"))
	v.SetCell(row+1, 0, v.sectionCell("MEM"))
	v.SetCell(row+1, 1, v.infoCell("n/a"))
	v.refresh()
}

func (*infoView) sectionCell(t string) *tview.TableCell {
	c := tview.NewTableCell(t + ":")
	c.SetAlign(tview.AlignLeft)
	return c
}

func (*infoView) infoCell(t string) *tview.TableCell {
	c := tview.NewTableCell(t)
	c.SetExpansion(2)
	c.SetTextColor(tcell.ColorOrange)
	return c
}

func (v *infoView) refresh() {
	var row int

	cluster := resource.NewCluster()
	v.GetCell(row, 1).SetText(cluster.ContextName())
	row++
	v.GetCell(row, 1).SetText(cluster.ClusterName())
	row++
	v.GetCell(row, 1).SetText(cluster.UserName())
	row += 2
	v.GetCell(row, 1).SetText(cluster.Version())
	row++

	mx, err := cluster.Metrics()
	if err != nil {
		log.Error().Err(err)
		return
	}
	c := v.GetCell(row, 1)
	c.SetText(deltas(c.Text, mx.CPU))
	row++
	c = v.GetCell(row, 1)
	c.SetText(deltas(c.Text, mx.Mem))
}
