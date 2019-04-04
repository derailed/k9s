package views

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

type selectList struct {
	*tview.List

	parent  loggable
	actions keyActions
}

func newSelectList(parent loggable) *selectList {
	v := selectList{List: tview.NewList(), actions: keyActions{}}
	{
		v.parent = parent
		v.SetBorder(true)
		v.SetMainTextColor(tcell.ColorGray)
		v.ShowSecondaryText(false)
		v.SetShortcutColor(tcell.ColorAqua)
		v.SetSelectedBackgroundColor(tcell.ColorAqua)
		v.SetTitle(" [aqua::b]Container Selector ")
		v.SetInputCapture(func(evt *tcell.EventKey) *tcell.EventKey {
			if a, ok := v.actions[evt.Key()]; ok {
				a.action(evt)
				evt = nil
			}
			return evt
		})
	}

	return &v
}

func (v *selectList) back(evt *tcell.EventKey) *tcell.EventKey {
	v.parent.switchPage(v.parent.getList().GetName())

	return nil
}

// Protocol...

func (v *selectList) switchPage(p string) {
	v.parent.switchPage(p)
}

func (v *selectList) backFn() actionHandler {
	return v.parent.backFn()
}

func (v *selectList) appView() *appView {
	return v.parent.appView()
}

func (v *selectList) getList() resource.List {
	return v.parent.getList()
}

func (v *selectList) getSelection() string {
	return v.parent.getSelection()
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
		v.AddItem(s, "Select a container", rune('a'+i), nil)
	}
}
