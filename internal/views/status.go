package views

import (
	"fmt"

	"github.com/derailed/tview"
)

type statusView struct {
	*tview.TextView

	app *appView
}

func newStatusView(app *appView) *statusView {
	v := statusView{app: app, TextView: tview.NewTextView()}
	{
		v.SetBackgroundColor(app.styles.BgColor())
		v.SetTextAlign(tview.AlignRight)
		// v.SetBorderPadding(0, 0, 1, 1)
		v.SetDynamicColors(true)
	}
	return &v
}

func (v *statusView) update(status []string) {
	v.Clear()
	last, bgColor := len(status)-1, v.app.styles.Style.Crumb.BgColor
	for i, c := range status {
		if i == last {
			bgColor = v.app.styles.Style.Crumb.ActiveColor
		}
		fmt.Fprintf(v, "[%s:%s:b] %s [-:%s:-] ",
			v.app.styles.Style.Crumb.FgColor,
			bgColor, c,
			v.app.styles.Style.BgColor)
	}
}
