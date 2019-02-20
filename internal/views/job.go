package views

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	log "github.com/sirupsen/logrus"
)

type jobView struct {
	*resourceView
}

func newJobView(t string, app *appView, list resource.List, c colorerFn) resourceViewer {
	v := jobView{newResourceView(t, app, list, c).(*resourceView)}
	v.extraActionsFn = v.extraActions
	v.AddPage("logs", newLogsView(&v), true, false)
	v.switchPage("job")
	return &v
}

// Protocol...

func (v *jobView) appView() *appView {
	return v.app
}
func (v *jobView) getList() resource.List {
	return v.list
}
func (v *jobView) getSelection() string {
	return v.selectedItem
}

// Handlers...

func (v *jobView) logs(*tcell.EventKey) {
	if !v.rowSelected() {
		return
	}

	cc, err := fetchContainers(v.list, v.selectedItem)
	if err != nil {
		log.Error(err)
	}

	l := v.GetPrimitive("logs").(*logsView)
	l.deleteAllPages()
	for _, c := range cc {
		l.addContainer(c)
	}

	v.switchPage("logs")
	l.init()
}

func (v *jobView) extraActions(aa keyActions) {
	aa[KeyL] = newKeyHandler("Logs", v.logs)
}
