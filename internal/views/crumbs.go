package views

import (
	"fmt"

	"github.com/derailed/tview"
)

type crumbsView struct {
	*tview.TextView

	app *appView
}

func newCrumbsView(app *appView) *crumbsView {
	v := crumbsView{app: app, TextView: tview.NewTextView()}
	{
		v.SetBackgroundColor(app.styles.BgColor())
		v.SetTextAlign(tview.AlignLeft)
		v.SetBorderPadding(0, 0, 1, 1)
		v.SetDynamicColors(true)
	}

	return &v
}

func (v *crumbsView) update(crumbs []string) {
	v.Clear()
	last, bgColor := len(crumbs)-1, v.app.styles.Style.Crumb.BgColor
	for i, c := range crumbs {
		if i == last {
			bgColor = v.app.styles.Style.Crumb.ActiveColor
		}
		fmt.Fprintf(v, "[%s:%s:b] <%s> [-:%s:-] ",
			v.app.styles.Style.Crumb.FgColor,
			bgColor, c,
			v.app.styles.Style.BgColor)
	}
}
