package views

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
)

type crumbsView struct {
	*tview.TextView

	styles *config.Styles
}

func newCrumbsView(styles *config.Styles) *crumbsView {
	v := crumbsView{styles: styles, TextView: tview.NewTextView()}
	{
		v.SetBackgroundColor(styles.BgColor())
		v.SetTextAlign(tview.AlignLeft)
		v.SetBorderPadding(0, 0, 1, 1)
		v.SetDynamicColors(true)
	}

	return &v
}

func (v *crumbsView) update(crumbs []string) {
	v.Clear()
	last, bgColor := len(crumbs)-1, v.styles.Style.Crumb.BgColor
	for i, c := range crumbs {
		if i == last {
			bgColor = v.styles.Style.Crumb.ActiveColor
		}
		fmt.Fprintf(v, "[%s:%s:b] <%s> [-:%s:-] ",
			v.styles.Style.Crumb.FgColor,
			bgColor, c,
			v.styles.Style.BgColor)
	}
}
