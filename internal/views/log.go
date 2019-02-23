package views

import (
	"fmt"

	log "github.com/sirupsen/logrus"
)

type logView struct {
	*detailsView
}

func newLogView(title string, parent loggable) *logView {
	log.Debug("LogsView init...")
	v := logView{detailsView: newDetailsView(parent.appView(), parent.backFn())}
	{
		v.SetBorderPadding(0, 0, 1, 1)
		v.setCategory("Logs")
		v.setTitle(parent.getSelection())
	}
	return &v
}

func (l *logView) log(lines fmt.Stringer) {
	l.Clear()
	fmt.Fprintln(l, lines.String())
}
