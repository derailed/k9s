package views

import (
	"fmt"
	"io"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
)

type logView struct {
	*detailsView
	ansiWriter io.Writer
}

func newLogView(title string, parent loggable) *logView {
	v := logView{detailsView: newDetailsView(parent.appView(), parent.backFn())}
	{
		v.SetBorderPadding(0, 0, 1, 1)
		v.setCategory("Logs")
		v.SetDynamicColors(true)
		v.SetWrap(true)
		v.setTitle(parent.getSelection())
		v.SetMaxBuffer(config.Root.K9s.LogBufferSize)
	}
	v.ansiWriter = tview.ANSIWriter(v)
	return &v
}

func (l *logView) logLine(line string) {
	fmt.Fprintln(l.ansiWriter, line)
}

func (l *logView) log(lines fmt.Stringer) {
	l.Clear()
	fmt.Fprintln(l.ansiWriter, lines.String())
}
