package views

import (
	"context"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	maxBuff1     int64 = 200
	refreshRate        = 200 * time.Millisecond
	maxCleanse         = 100
	logBuffSize        = 100
	flushTimeout       = 200 * time.Millisecond

	logCoFmt = " Logs([fg:bg:]%s:[hilite:bg:b]%s[-:bg:-]) "
	logFmt   = " Logs([fg:bg:]%s) "
)

type (
	masterView interface {
		backFn() actionHandler
		appView() *appView
	}

	logsView struct {
		*tview.Pages

		app        *appView
		parent     loggable
		actions    keyActions
		cancelFunc context.CancelFunc
	}
)

func newLogsView(title string, app *appView, parent loggable) *logsView {
	v := logsView{
		app:    app,
		Pages:  tview.NewPages(),
		parent: parent,
	}

	return &v
}

// Protocol...

func (v *logsView) reload(co string, parent loggable, prevLogs bool) {
	v.parent = parent
	v.deletePage()
	v.AddPage("logs", newLogView(co, v.app, v.backCmd), true, true)
	v.load(co, prevLogs)
}

// SetActions to handle keyboard events.
func (v *logsView) setActions(aa keyActions) {
	v.actions = aa
}

// Hints show action hints
func (v *logsView) hints() hints {
	l := v.CurrentPage().Item.(*logView)
	return l.actions.toHints()
}

func (v *logsView) backFn() actionHandler {
	return v.backCmd
}

func (v *logsView) deletePage() {
	v.RemovePage("logs")
}

func (v *logsView) stop() {
	if v.cancelFunc == nil {
		return
	}
	v.cancelFunc()
	log.Debug().Msgf("Canceling logs...")
	v.cancelFunc = nil
}

func (v *logsView) load(container string, prevLogs bool) {
	if err := v.doLoad(v.parent.getSelection(), container, prevLogs); err != nil {
		v.app.flash().err(err)
		l := v.CurrentPage().Item.(*logView)
		l.logLine("😂 Doh! No logs are available at this time. Check again later on...")
		return
	}
	v.app.SetFocus(v)
}

func (v *logsView) doLoad(path, co string, prevLogs bool) error {
	v.stop()

	l := v.CurrentPage().Item.(*logView)
	l.logs.Clear()
	l.path = path
	l.setTitle(path, co)

	c := make(chan string, 10)
	go updateLogs(c, l)

	res, ok := v.parent.getList().Resource().(resource.Tailable)
	if !ok {
		return fmt.Errorf("Resource %T is not tailable", v.parent.getList().Resource())
	}

	var ctx context.Context
	ctx = context.WithValue(context.Background(), resource.IKey("informer"), v.app.informer)
	ctx, v.cancelFunc = context.WithCancel(ctx)
	if err := res.Logs(ctx, c, v.logOpts(path, co, prevLogs)); err != nil {
		v.cancelFunc()
		return err
	}

	return nil
}

func (v *logsView) logOpts(path, co string, prevLogs bool) resource.LogOptions {
	ns, po := namespaced(path)
	return resource.LogOptions{
		Namespace: ns,
		Name:      po,
		Container: co,
		Lines:     int64(v.app.config.K9s.LogRequestSize),
		Previous:  prevLogs,
	}

}
func updateLogs(c <-chan string, l *logView) {
	buff, index := make([]string, logBuffSize), 0
	for {
		select {
		case line, ok := <-c:
			if !ok {
				l.flush(index, buff)
				index = 0
				return
			}
			if index < logBuffSize {
				buff[index] = line
				index++
				continue
			}
			l.flush(index, buff)
			index = 0
			buff[index] = line
		case <-time.After(flushTimeout):
			l.flush(index, buff)
			index = 0
		}
	}
}

// ----------------------------------------------------------------------------
// Actions...

func (v *logsView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.stop()
	v.parent.switchPage("master")

	return evt
}
