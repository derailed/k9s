package views

import (
	"strconv"

	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
)

type selectList struct {
	*tview.List

	actions keyActions
}

func newSelectList() *selectList {
	v := selectList{List: tview.NewList()}
	v.SetBorder(true)
	v.SetTitle(" Please select a Container ")
	v.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
		if a, ok := v.actions[evt.Key()]; ok {
			a.action(evt)
			evt = nil
		}
		return evt
	})
	return &v
}

// SetActions to handle keyboard events.
func (v *selectList) setActions(aa keyActions) {
	v.actions = aa
}

func (v *selectList) hints() hints {
	if v.actions != nil {
		return v.actions.toHints()
	}
	return nil
}

func (v *selectList) populate(ss []string) {
	v.Clear()
	for i, s := range ss {
		v.AddItem(s, "Select a container", rune(strconv.Itoa(i)[0]), nil)
	}
}
