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

	logFmt = " Logs([fg:bg:]%s:[hilite:bg:b]%s[-:bg:-]) "
)

type masterView interface {
	backFn() actionHandler
	appView() *appView
}

type logsView struct {
	*tview.Pages

	parentView   string
	parent       loggable
	containers   []string
	actions      keyActions
	cancelFunc   context.CancelFunc
	showPrevious bool
}

func newLogsView(pview string, parent loggable) *logsView {
	v := logsView{
		Pages:      tview.NewPages(),
		parent:     parent,
		parentView: pview,
		containers: []string{},
	}

	return &v
}

// Protocol...

func (v *logsView) reload(co string, parent loggable, view string, prevLogs bool) {
	v.parent, v.parentView, v.showPrevious = parent, view, prevLogs
	v.deleteAllPages()
	v.addContainer(co)
	v.load(0)
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

func (v *logsView) addContainer(n string) {
	v.containers = append(v.containers, n)
	l := newLogView(n, v)
	v.AddPage(n, l, true, false)
}

func (v *logsView) appView() *appView {
	return v.parent.appView()
}

func (v *logsView) backFn() actionHandler {
	return v.backCmd
}

func (v *logsView) deleteAllPages() {
	for i, c := range v.containers {
		v.RemovePage(c)
		delete(v.actions, tcell.Key(numKeys[i+1]))
	}
	v.containers = []string{}
}

func (v *logsView) stop() {
	if v.cancelFunc == nil {
		return
	}
	v.cancelFunc()
	log.Debug().Msgf("Canceling logs...")
	v.cancelFunc = nil
}

func (v *logsView) load(i int) {
	if i < 0 || i > len(v.containers)-1 {
		return
	}
	v.SwitchToPage(v.containers[i])
	if err := v.doLoad(v.parent.getSelection(), v.containers[i]); err != nil {
		v.parent.appView().flash().err(err)
		l := v.CurrentPage().Item.(*logView)
		l.logLine("😂 Doh! No logs are available at this time. Check again later on...")
		return
	}
	v.parent.appView().SetFocus(v)
}

func (v *logsView) doLoad(path, co string) error {
	v.stop()

	maxBuff := int64(v.parent.appView().config.K9s.LogRequestSize)
	l := v.CurrentPage().Item.(*logView)
	l.logs.Clear()
	fmat := skinTitle(fmt.Sprintf(logFmt, path, co), v.parent.appView().styles.Style)
	l.SetTitle(fmat)

	c := make(chan string, 10)
	go func(l *logView) {
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
	}(l)

	ns, po := namespaced(path)
	res, ok := v.parent.getList().Resource().(resource.Tailable)
	if !ok {
		return fmt.Errorf("Resource %T is not tailable", v.parent.getList().Resource)
	}

	cancelFn, err := res.Logs(c, ns, po, co, maxBuff, v.showPrevious)
	if err != nil {
		cancelFn()
		return err
	}
	v.cancelFunc = cancelFn

	return nil
}

// ----------------------------------------------------------------------------
// Actions...

func (v *logsView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.stop()
	v.parent.switchPage(v.parentView)

	return evt
}
