package views

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	log "github.com/sirupsen/logrus"
)

const (
	maxBuff1    int64 = 200
	refreshRate       = 200 * time.Millisecond
	maxCleanse        = 100
)

type logsView struct {
	*tview.Pages

	parent     loggable
	containers []string
	actions    keyActions
	cancelFunc context.CancelFunc
	buffer     *logBuffer
}

func newLogsView(parent loggable) *logsView {
	maxBuff := config.Root.K9s.LogBufferSize
	v := logsView{
		Pages:      tview.NewPages(),
		parent:     parent,
		containers: []string{},
		buffer:     newLogBuffer(int(maxBuff), true),
	}
	v.setActions(keyActions{
		tcell.KeyEscape: {description: "Back", action: v.back},
		KeyC:            {description: "Clear", action: v.clearLogs},
		KeyU:            {description: "Top", action: v.top},
		KeyD:            {description: "Bottom", action: v.bottom},
		KeyF:            {description: "Up", action: v.pageUp},
		KeyB:            {description: "Down", action: v.pageDown},
	})
	v.SetInputCapture(v.keyboard)

	return &v
}

// Protocol...

func (v *logsView) init() {
	v.load(0)
}

func (v *logsView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	if kv, ok := v.CurrentPage().Item.(keyHandler); ok {
		if kv.keyboard(evt) == nil {
			return nil
		}
	}

	key := evt.Key()
	if key == tcell.KeyRune {
		if i, err := strconv.Atoi(string(evt.Rune())); err == nil {
			if _, ok := numKeys[i]; ok {
				v.load(i - 1)
				return nil
			}
		}
		key = tcell.Key(evt.Rune())
	}

	if m, ok := v.actions[key]; ok {
		log.Debug(">> LogsView handled ", tcell.KeyNames[key])
		return m.action(evt)
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
			v.actions[tcell.Key(numKeys[i+1])] = newKeyAction(c, nil)
		}
	}
	return v.actions.toHints()
}

func (v *logsView) addContainer(n string) {
	v.containers = append(v.containers, n)
	l := newLogView(n, v.parent)
	{
		l.SetInputCapture(v.keyboard)
	}
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
	if err := v.doLoad(v.parent.getSelection(), v.containers[i]); err != nil {
		v.parent.appView().flash(flashErr, err.Error())
		v.buffer.add("😂 Doh! No logs are available at this time. Check again later on...")
		l := v.CurrentPage().Item.(*logView)
		l.log(v.buffer)
		return
	}
	v.parent.appView().SetFocus(v)
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
		l.setTitle(path + ":" + co)
		for {
			select {
			case line, ok := <-c:
				if !ok {
					if v.buffer.length() > 0 {
						v.buffer.add("--- No more logs ---")
						l.log(v.buffer)
						l.ScrollToEnd()
					}
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
	res, ok := v.parent.getList().Resource().(resource.Tailable)
	if !ok {
		return fmt.Errorf("Resource %T is not tailable", v.parent.getList().Resource)
	}
	maxBuff := config.Root.K9s.LogBufferSize
	cancelFn, err := res.Logs(c, ns, po, co, int64(maxBuff), false)
	if err != nil {
		cancelFn()
		return err
	}
	v.cancelFunc = cancelFn
	return nil
}

// ----------------------------------------------------------------------------
// Actions...

func (v *logsView) back(evt *tcell.EventKey) *tcell.EventKey {
	v.stop()
	v.parent.switchPage(v.parent.getList().GetName())
	return nil
}

func (v *logsView) top(evt *tcell.EventKey) *tcell.EventKey {
	if p := v.CurrentPage(); p != nil {
		v.parent.appView().flash(flashInfo, "Top logs...")
		p.Item.(*logView).ScrollToBeginning()
	}
	return nil
}

func (v *logsView) bottom(*tcell.EventKey) *tcell.EventKey {
	if p := v.CurrentPage(); p != nil {
		v.parent.appView().flash(flashInfo, "Bottom logs...")
		p.Item.(*logView).ScrollToEnd()
	}
	return nil
}

func (v *logsView) pageUp(*tcell.EventKey) *tcell.EventKey {
	if p := v.CurrentPage(); p != nil {
		v.parent.appView().flash(flashInfo, "Page Up logs...")
		p.Item.(*logView).PageUp()
	}
	return nil
}

func (v *logsView) pageDown(*tcell.EventKey) *tcell.EventKey {
	if p := v.CurrentPage(); p != nil {
		v.parent.appView().flash(flashInfo, "Page Down logs...")
		p.Item.(*logView).PageDown()
	}
	return nil
}

func (v *logsView) clearLogs(*tcell.EventKey) *tcell.EventKey {
	if p := v.CurrentPage(); p != nil {
		v.parent.appView().flash(flashInfo, "Clearing logs...")
		v.buffer.clear()
		p.Item.(*logView).Clear()
	}
	return nil
}
