package views

import (
	"fmt"
)

type logView struct {
	*detailsView
}

func newLogView(title string, parent loggable) *logView {
	v := logView{detailsView: newDetailsView(parent.appView(), parent.backFn())}
	{
		v.SetBorderPadding(0, 0, 1, 1)
		v.setCategory("Logs")
		v.SetDynamicColors(false)
		v.SetWrap(true)
		v.setTitle(parent.getSelection())
	}
	return &v
}

func (l *logView) log(lines fmt.Stringer) {
	l.Clear()
	fmt.Fprintln(l, lines.String())
}
