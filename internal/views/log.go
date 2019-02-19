package views

import (
	"fmt"

	"github.com/k8sland/tview"
)

const (
	logTitleFmt = " [aqua::b]Logs %s ([fuchsia::-]container=[fuchsia::b]%s[aqua::b]) "
)

type logView struct {
	*tview.TextView
}

func newLogView(title string, pv *podView) *logView {
	v := logView{TextView: tview.NewTextView()}
	{
		v.SetScrollable(true)
		v.SetDynamicColors(true)
		v.SetBorder(true)
		v.SetBorderPadding(0, 0, 1, 1)
		v.SetTitle(fmt.Sprintf(logTitleFmt, pv.selectedItem, title))
		v.SetWrap(false)
		v.SetChangedFunc(func() {
			pv.app.Draw()
		})
	}
	return &v
}

func (l *logView) log(lines fmt.Stringer) {
	l.Clear()
	fmt.Fprintln(l, lines.String())
}
