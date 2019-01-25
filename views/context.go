package views

import (
	"github.com/gdamore/tcell"
	"github.com/k8sland/k9s/resource"
)

type contextView struct {
	*resourceView
}

func newContextView(t string, app *appView, list resource.List, c colorerFn) resourceViewer {
	v := contextView{newResourceView(t, app, list, c).(*resourceView)}
	v.extraActionsFn = v.extraActions

	v.switchPage("ctx")

	return &v
}

func (v *contextView) useContext(*tcell.EventKey) {
	if !v.rowSelected() {
		return
	}
	err := v.list.Resource().(*resource.Context).Switch(v.selectedItem)
	if err != nil {
		v.app.flash(flashWarn, err.Error())
		return
	}
	v.app.flash(flashInfo, "Switching context to ", v.selectedItem)
	v.refresh()
	table := v.GetPrimitive("ctx").(*tableView)
	table.Select(0, 0)
}

func (v *contextView) extraActions(aa keyActions) {
	aa[tcell.KeyCtrlS] = keyAction{description: "Switch", action: v.useContext}
}
