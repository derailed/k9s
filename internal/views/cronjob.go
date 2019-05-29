package views

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
)

type cronJobView struct {
	*resourceView
}

func newCronJobView(t string, app *appView, list resource.List) resourceViewer {
	v := cronJobView{resourceView: newResourceView(t, app, list).(*resourceView)}
	v.extraActionsFn = v.extraActions

	return &v
}

func (v *cronJobView) trigger(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	if err := v.list.Resource().(resource.Runner).Run(v.selectedItem); err != nil {
		v.app.flash().errf("Cronjob trigger failed %v", err)
		return evt
	}
	v.app.flash().infof("Triggering %s %s", v.list.GetName(), v.selectedItem)

	return nil
}

func (v *cronJobView) extraActions(aa keyActions) {
	aa[tcell.KeyCtrlT] = newKeyAction("Trigger", v.trigger, true)
}
