package views

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

type jobView struct {
	*resourceView
}

func newJobView(t string, app *appView, list resource.List, c colorerFn) resourceViewer {
	v := jobView{newResourceView(t, app, list, c).(*resourceView)}
	{
		v.extraActionsFn = v.extraActions
	}
	v.AddPage("logs", newLogsView(&v), true, false)
	v.switchPage("job")
	return &v
}

// Protocol...

func (v *jobView) appView() *appView {
	return v.app
}

func (v *jobView) backFn() actionHandler {
	return v.backCmd
}

func (v *jobView) getList() resource.List {
	return v.list
}

func (v *jobView) getSelection() string {
	return v.selectedItem
}

// Handlers...

func (v *jobView) logs(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	cc, err := fetchContainers(v.list, v.selectedItem, true)
	if err != nil {
		v.app.flash(flashErr, err.Error())
		log.Error().Err(err)
		return evt
	}

	l := v.GetPrimitive("logs").(*logsView)
	l.deleteAllPages()
	for _, c := range cc {
		l.addContainer(c)
	}

	v.switchPage("logs")
	l.init()
	return nil
}

func (v *jobView) extraActions(aa keyActions) {
	aa[KeyL] = newKeyAction("Logs", v.logs)
}
