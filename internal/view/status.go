package view

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
)

type statusView struct {
	*tview.TextView

	styles *config.Styles
}

func newStatusView(styles *config.Styles) *statusView {
	v := statusView{styles: styles, TextView: tview.NewTextView()}
	{
		v.SetBackgroundColor(config.AsColor(styles.Views().Log.BgColor))
		v.SetTextAlign(tview.AlignRight)
		v.SetDynamicColors(true)
	}
	return &v
}

func (v *statusView) update(status []string) {
	v.Clear()
	last, bgColor := len(status)-1, v.styles.Frame().Crumb.BgColor
	for i, c := range status {
		if i == last {
			bgColor = v.styles.Frame().Crumb.ActiveColor
		}
		fmt.Fprintf(v, "[%s:%s:b] %-15s ", v.styles.Frame().Crumb.FgColor, bgColor, c)
	}
}
