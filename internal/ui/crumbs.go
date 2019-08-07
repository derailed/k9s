package ui

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
)

// CrumbsView represents user breadcrumbs.
type CrumbsView struct {
	*tview.TextView

	styles *config.Styles
}

// NewCrumbsView returns a new breadcrumb view.
func NewCrumbsView(styles *config.Styles) *CrumbsView {
	v := CrumbsView{styles: styles, TextView: tview.NewTextView()}
	{
		v.SetBackgroundColor(styles.BgColor())
		v.SetTextAlign(tview.AlignLeft)
		v.SetBorderPadding(0, 0, 1, 1)
		v.SetDynamicColors(true)
	}

	return &v
}

// Refresh updates view with new crumbs.
func (v *CrumbsView) Refresh(crumbs []string) {
	v.Clear()
	last, bgColor := len(crumbs)-1, v.styles.Frame().Crumb.BgColor
	for i, c := range crumbs {
		if i == last {
			bgColor = v.styles.Frame().Crumb.ActiveColor
		}
		fmt.Fprintf(v, "[%s:%s:b] <%s> [-:%s:-] ",
			v.styles.Frame().Crumb.FgColor,
			bgColor, c,
			v.styles.Body().BgColor)
	}
}
