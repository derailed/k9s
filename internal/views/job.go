package views

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

type jobView struct {
	*resourceView
}

func newJobView(t string, app *appView, list resource.List) resourceViewer {
	v := jobView{resourceView: newResourceView(t, app, list).(*resourceView)}
	{
		v.extraActionsFn = v.extraActions
		v.AddPage("logs", newLogsView(list.GetName(), &v), true, false)
		v.switchPage("job")
	}

	picker := newSelectList(&v)
	{
		picker.setActions(keyActions{
			tcell.KeyEscape: {description: "Back", action: v.backCmd, visible: true},
		})
	}
	v.AddPage("picker", picker, true, false)

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

func (v *jobView) logsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.viewLogs(false) {
		return nil
	}
	return evt
}

func (v *jobView) prevLogsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.viewLogs(true) {
		return nil
	}
	return evt
}

func (v *jobView) viewLogs(previous bool) bool {
	if !v.rowSelected() {
		return false
	}

	cc, err := fetchContainers(v.list, v.selectedItem, true)
	if err != nil {
		v.app.flash(flashErr, err.Error())
		log.Error().Err(err).Msgf("Unable to fetch containers for %s", v.selectedItem)
		return false
	}

	if len(cc) == 1 {
		v.showLogs(v.selectedItem, cc[0], v.list.GetName(), v, previous)
		return true
	}

	picker := v.GetPrimitive("picker").(*selectList)
	picker.populate(cc)
	picker.SetSelectedFunc(func(i int, t, d string, r rune) {
		v.showLogs(v.selectedItem, t, "picker", picker, previous)
	})
	v.switchPage("picker")

	return true
}

func (v *jobView) showLogs(path, co, view string, parent loggable, prev bool) {
	l := v.GetPrimitive("logs").(*logsView)
	l.reload(co, parent, view, prev)
	v.switchPage("logs")
}

func (v *jobView) extraActions(aa keyActions) {
	aa[KeyL] = newKeyAction("Logs", v.logsCmd, true)
	aa[KeyShiftL] = newKeyAction("Previous Logs", v.prevLogsCmd, true)
}
