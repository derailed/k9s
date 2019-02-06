package views

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/derailed/k9s/resource"
	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
	log "github.com/sirupsen/logrus"
)

const (
	maxBuff1    int64 = 200
	refreshRate       = 200 * time.Millisecond
	maxCleanse        = 100
)

type logsView struct {
	*tview.Pages

	pv         *podView
	containers []string
	actions    keyActions
	cancelFunc context.CancelFunc
	buffer     *logBuffer
}

func newLogsView(pv *podView) *logsView {
	maxBuff := k9sCfg.K9s.LogBufferSize
	v := logsView{
		Pages:      tview.NewPages(),
		pv:         pv,
		containers: []string{},
		buffer:     newLogBuffer(int(maxBuff), true),
	}
	v.setActions(keyActions{
		tcell.KeyEscape: {description: "Back", action: v.back},
		tcell.KeyCtrlK:  {description: "Clear", action: v.clearLogs},
		tcell.KeyCtrlU:  {description: "Top", action: v.top},
		tcell.KeyCtrlD:  {description: "Bottom", action: v.bottom},
		tcell.KeyCtrlF:  {description: "Page Up", action: v.pageUp},
		tcell.KeyCtrlB:  {description: "Page Down", action: v.pageDown},
	})
	v.SetInputCapture(v.keyboard)

	return &v
}

// Protocol...

func (v *logsView) init() {
	v.load(0)
}

func (v *logsView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	if m, ok := v.actions[evt.Key()]; ok {
		m.action(evt)
		return nil
	}

	if evt.Key() != tcell.KeyRune {
		return evt
	}

	i, err := strconv.Atoi(string(evt.Rune()))
	if err != nil {
		log.Error("Boom!", err)
		return evt
	}
	if _, ok := numKeys[i]; ok {
		v.load(i - 1)
		v.pv.app.resetCmd()
		return nil
	}
	return evt
}

// SetActions to handle keyboard events.
func (v *logsView) setActions(aa keyActions) {
	v.actions = aa
}

// Hints show action hints
func (v *logsView) hints() hints {
	if len(v.containers) > 1 {
		for i, c := range v.containers {
			v.actions[tcell.Key(numKeys[i+1])] = keyAction{description: c}
		}
	}
	return v.actions.toHints()
}

func (v *logsView) addContainer(n string) {
	v.containers = append(v.containers, n)
	l := newLogView(n, v.pv)
	l.SetInputCapture(v.keyboard)
	v.AddPage(n, l, true, false)
}

func (v *logsView) deleteAllPages() {
	for i, c := range v.containers {
		v.RemovePage(c)
		delete(v.actions, tcell.Key(numKeys[i+1]))
	}
	v.containers = []string{}
}

func (v *logsView) stop() {
	v.killLogIfAny()
}

func (v *logsView) load(i int) {
	if i < 0 || i > len(v.containers)-1 {
		return
	}
	v.SwitchToPage(v.containers[i])
	v.buffer.clear()
	if err := v.doLoad(v.pv.selectedItem, v.containers[i]); err != nil {
		v.pv.app.flash(flashErr, err.Error())
		v.buffer.add("😂 Doh! No logs are available at this time. Check again later on...")
		l := v.CurrentPage().Item.(*logView)
		l.log(v.buffer)
		return
	}
	v.pv.app.SetFocus(v)
}

func (v *logsView) killLogIfAny() {
	if v.cancelFunc == nil {
		return
	}
	v.cancelFunc()
	v.cancelFunc = nil
}

func (v *logsView) doLoad(path, co string) error {
	v.killLogIfAny()

	c := make(chan string)
	go func() {
		l, count, first := v.CurrentPage().Item.(*logView), 0, true
		for {
			select {
			case line, ok := <-c:
				if !ok {
					return
				}
				v.buffer.add(line)
			case <-time.After(refreshRate):
				if count == maxCleanse {
					log.Debug("Cleansing logs")
					v.buffer.cleanse()
					count = 0
				}
				count++
				if v.buffer.length() == 0 {
					l.Clear()
					continue
				}
				l.log(v.buffer)
				if first {
					l.ScrollToEnd()
					first = false
				}
			}
		}
	}()

	ns, po := namespaced(path)
	res, ok := v.pv.list.Resource().(resource.Tailable)
	if !ok {
		return fmt.Errorf("Resource %T is not tailable", v.pv.list.Resource)
	}
	maxBuff := k9sCfg.K9s.LogBufferSize
	cancelFn, err := res.Logs(c, ns, po, co, int64(maxBuff))
	if err != nil {
		cancelFn()
		return err
	}
	v.cancelFunc = cancelFn

	return nil
}

// ----------------------------------------------------------------------------
// Actions...

func (v *logsView) back(*tcell.EventKey) {
	v.stop()
	v.pv.switchPage(v.pv.list.GetName())
}

func (v *logsView) top(*tcell.EventKey) {
	if p := v.CurrentPage(); p != nil {
		v.pv.app.flash(flashInfo, "Top logs...")
		p.Item.(*logView).ScrollToBeginning()
	}
}

func (v *logsView) bottom(*tcell.EventKey) {
	if p := v.CurrentPage(); p != nil {
		v.pv.app.flash(flashInfo, "Bottom logs...")
		p.Item.(*logView).ScrollToEnd()
	}
}

func (v *logsView) pageUp(*tcell.EventKey) {
	if p := v.CurrentPage(); p != nil {
		v.pv.app.flash(flashInfo, "Page Up logs...")
		p.Item.(*logView).PageUp()
	}
}

func (v *logsView) pageDown(*tcell.EventKey) {
	if p := v.CurrentPage(); p != nil {
		v.pv.app.flash(flashInfo, "Page Down logs...")
		p.Item.(*logView).PageDown()
	}
}

func (v *logsView) clearLogs(*tcell.EventKey) {
	if p := v.CurrentPage(); p != nil {
		v.pv.app.flash(flashInfo, "Clearing logs...")
		v.buffer.clear()
		p.Item.(*logView).Clear()
	}
}
