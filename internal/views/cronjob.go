package views

import (
	"fmt"
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
)

type cronjobView struct {
	*resourceView
}

func newCronjobView(t string, app *appView, list resource.List, c colorerFn) resourceViewer {
	v := cronjobView{
		resourceView: newResourceView(t, app, list, c).(*resourceView),
	}
	v.extraActionsFn = v.extraActions
	v.switchPage("job")
	return &v
}

func (v *cronjobView) trigger(*tcell.EventKey) {
	if !v.rowSelected() {
		return
	}

	v.app.flash(flashInfo, fmt.Sprintf("Triggering %s %s", v.list.GetName(), v.selectedItem))
	if err := v.list.Resource().(resource.TriggerableCronjob).Trigger(v.selectedItem); err != nil {
		v.app.flash(flashErr, "Boom!", err.Error())
	}
}

func (v *cronjobView) extraActions(aa keyActions) {
	aa[tcell.KeyCtrlT] = keyAction{description: "Trigger", action: v.trigger}
}
