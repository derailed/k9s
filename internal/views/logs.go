package views

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
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
}

func newLogsView(parent loggable) *logsView {
	v := logsView{
		Pages:      tview.NewPages(),
		parent:     parent,
		containers: []string{},
	}
	v.setActions(keyActions{
		tcell.KeyEscape: {description: "Back", action: v.back, visible: true},
		KeyC:            {description: "Clear", action: v.clearLogs, visible: true},
		KeyG:            {description: "Top", action: v.top, visible: false},
		KeyShiftG:       {description: "Bottom", action: v.bottom, visible: false},
		KeyF:            {description: "Up", action: v.pageUp, visible: false},
		KeyB:            {description: "Down", action: v.pageDown, visible: false},
	})
	v.SetInputCapture(v.keyboard)

	return &v
}

// Protocol...

func (v *logsView) init() {
	v.load(0)
}

func (v *logsView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		key = tcell.Key(evt.Rune())
	}

	if kv, ok := v.CurrentPage().Item.(keyHandler); ok {
		if kv.keyboard(evt) == nil {
			return nil
		}
	}

	if evt.Key() == tcell.KeyRune {
		if i, err := strconv.Atoi(string(evt.Rune())); err == nil {
			if _, ok := numKeys[i]; ok {
				v.load(i - 1)
				return nil
			}
		}
	}

	if m, ok := v.actions[key]; ok {
		log.Debug().Msgf(">> LogsView handled %s", tcell.KeyNames[key])
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
			v.actions[tcell.Key(numKeys[i+1])] = newKeyAction(c, nil, true)
		}
	}
	return v.actions.toHints()
}

func (v *logsView) addContainer(n string) {
	v.containers = append(v.containers, n)
	l := newLogView(n, v.parent)
	{
		l.SetInputCapture(v.keyboard)
		l.backFn = v.back
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
	if v.cancelFunc == nil {
		return
	}
	log.Debug().Msg("Canceling logs...")
	v.cancelFunc()
	v.cancelFunc = nil
}

func (v *logsView) load(i int) {
	if i < 0 || i > len(v.containers)-1 {
		return
	}
	v.SwitchToPage(v.containers[i])
	if err := v.doLoad(v.parent.getSelection(), v.containers[i]); err != nil {
		v.parent.appView().flash(flashErr, err.Error())
		l := v.CurrentPage().Item.(*logView)
		l.logLine("😂 Doh! No logs are available at this time. Check again later on...")
		return
	}
	v.parent.appView().SetFocus(v)
}

func (v *logsView) doLoad(path, co string) error {
	v.stop()

	c := make(chan string)
	go func() {
		l := v.CurrentPage().Item.(*logView)
		l.Clear()
		l.setTitle(path + ":" + co)
		for {
			select {
			case line, ok := <-c:
				if !ok {
					l.ScrollToEnd()
					return
				}
				l.logLine(line)
			}
		}
	}()

	ns, po := namespaced(path)
	res, ok := v.parent.getList().Resource().(resource.Tailable)
	if !ok {
		return fmt.Errorf("Resource %T is not tailable", v.parent.getList().Resource)
	}
	maxBuff := int64(v.parent.appView().config.K9s.LogRequestSize)
	cancelFn, err := res.Logs(c, ns, po, co, maxBuff, false)
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
		v.parent.appView().flash(flashInfo, "Top of logs...")
		p.Item.(*logView).ScrollToBeginning()
	}
	return nil
}

func (v *logsView) bottom(*tcell.EventKey) *tcell.EventKey {
	if p := v.CurrentPage(); p != nil {
		v.parent.appView().flash(flashInfo, "Bottom of logs...")
		p.Item.(*logView).ScrollToEnd()
	}
	return nil
}

func (v *logsView) pageUp(*tcell.EventKey) *tcell.EventKey {
	if p := v.CurrentPage(); p != nil {
		if p.Item.(*logView).PageUp() {
			v.parent.appView().flash(flashInfo, "Reached Top ...")
		}
	}
	return nil
}

func (v *logsView) pageDown(*tcell.EventKey) *tcell.EventKey {
	if p := v.CurrentPage(); p != nil {
		if p.Item.(*logView).PageDown() {
			v.parent.appView().flash(flashInfo, "Reached Bottom ...")
		}
	}
	return nil
}

func (v *logsView) clearLogs(*tcell.EventKey) *tcell.EventKey {
	if p := v.CurrentPage(); p != nil {
		v.parent.appView().flash(flashInfo, "Clearing logs...")
		p.Item.(*logView).Clear()
	}
	return nil
}
