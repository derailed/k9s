package views

import (
	"fmt"

	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

type crumbsView struct {
	*tview.TextView

	app *tview.Application
}

func newCrumbsView(app *tview.Application) *crumbsView {
	v := crumbsView{app: app, TextView: tview.NewTextView()}
	{
		v.SetTextColor(tcell.ColorAqua)
		v.SetTextAlign(tview.AlignLeft)
		v.SetBorderPadding(0, 0, 1, 1)
		v.SetDynamicColors(true)
	}
	return &v
}

func (v *crumbsView) update(crumbs []string) {
	v.Clear()
	last, bgColor := len(crumbs)-1, "aqua"
	for i, c := range crumbs {
		if i == last {
			bgColor = "orange"
		}
		fmt.Fprintf(v, "[black:%s:b] <%s> [-:-:-]  ", bgColor, c)
	}
}
