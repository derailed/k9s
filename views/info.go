package views

import (
	"github.com/derailed/k9s/resource"
	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
	log "github.com/sirupsen/logrus"
)

type infoView struct {
	*tview.Table

	app *appView
}

func newInfoView(app *appView) *infoView {
	return &infoView{app: app, Table: tview.NewTable()}
}

func (v *infoView) init() {
	var row int

	cluster := resource.NewCluster()
	v.SetCell(row, 0, v.sectionCell("Cluster"))
	v.SetCell(row, 1, v.infoCell(cluster.Name()))
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
	v.GetCell(row, 1).SetText(cluster.Name())
	row+=2
	rev := cluster.Version()
	v.GetCell(row, 1).SetText(rev)
	row++

	mx, err := cluster.Metrics()
	if err != nil {
		log.Error(err)
		return
	}
	c := v.GetCell(row, 1)
	c.SetText(deltas(c.Text, mx.CPU))
	row++
	c = v.GetCell(row, 1)
	c.SetText(deltas(c.Text, mx.Mem))
}
