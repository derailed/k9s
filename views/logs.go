package views

import (
	"context"
	"fmt"
	"strconv"

	"github.com/gdamore/tcell"
	"github.com/k8sland/k9s/resource"
	"github.com/k8sland/tview"
	log "github.com/sirupsen/logrus"
)

type logsView struct {
	*tview.Pages

	pv         *podView
	containers []string
	actions    keyActions
	cancelFunc context.CancelFunc
}

func newLogsView(pv *podView) *logsView {
	v := logsView{Pages: tview.NewPages(), pv: pv, containers: []string{}}
	v.SetInputCapture(v.keyboard)

	return &v
}

func (v *logsView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	if m, ok := v.actions[evt.Key()]; ok {
		m.action(evt)
		return nil
	}

	if evt.Key() == tcell.KeyRune {
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

func (v *logsView) init() {
	v.load(0)
}

func (v *logsView) clearLogs() {
	p := v.CurrentPage()
	if p != nil {
		p.Item.(*logView).Clear()
	}
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
	if err := v.doLoad(v.pv.selectedItem, v.containers[i]); err != nil {
		v.pv.app.flash(flashErr, err.Error())
		return
	}
	v.pv.app.SetFocus(v)
}

func (v *logsView) killLogIfAny() {
	if v.cancelFunc != nil {
		v.cancelFunc()
		v.cancelFunc = nil
	}
}

func (v *logsView) doLoad(path, co string) error {
	v.killLogIfAny()

	c := make(chan string)
	go func() {
		l := v.CurrentPage().Item.(*logView)
		for s := range c {
			fmt.Fprintln(l, s)
		}
	}()

	ns, po := namespaced(path)
	cancelFn, err := v.pv.list.Resource().(resource.Tailable).Logs(c, ns, po, co)
	if err != nil {
		cancelFn()
		return err
	}
	v.cancelFunc = cancelFn

	return nil
}
