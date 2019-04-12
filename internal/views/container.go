package views

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

type containerView struct {
	*resourceView

	current igniter
	path    string
}

func newContainerView(t string, app *appView, list resource.List, path string) resourceViewer {
	v := containerView{resourceView: newResourceView(t, app, list).(*resourceView)}
	{
		v.path = path
		v.extraActionsFn = v.extraActions
		v.current = app.content.GetPrimitive("main").(igniter)
	}
	v.AddPage("logs", newLogsView(list.GetName(), &v), true, false)
	v.switchPage("co")

	return &v
}

// Protocol...

func (v *containerView) backFn() actionHandler {
	return v.backCmd
}

func (v *containerView) appView() *appView {
	return v.app
}

func (v *containerView) getList() resource.List {
	return v.list
}

func (v *containerView) getSelection() string {
	return v.path
}

// Handlers...

func (v *containerView) logsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	v.showLogs(v.selectedItem, v.list.GetName(), v, false)

	return nil
}

func (v *containerView) prevLogsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	v.showLogs(v.selectedItem, v.list.GetName(), v, true)

	return nil
}

func (v *containerView) showLogs(co, view string, parent loggable, prev bool) {
	l := v.GetPrimitive("logs").(*logsView)
	l.reload(co, parent, view, prev)
	v.switchPage("logs")
}

func (v *containerView) shellCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	log.Debug().Msgf("Selected %s", v.selectedItem)
	v.shellIn(v.path, v.selectedItem)

	return nil
}

func (v *containerView) shellIn(path, co string) {
	ns, po := namespaced(path)
	args := make([]string, 0, 12)
	args = append(args, "exec", "-it")
	args = append(args, "--context", v.app.config.K9s.CurrentContext)
	args = append(args, "-n", ns)
	args = append(args, po)
	if len(co) != 0 {
		args = append(args, "-c", co)
	}
	args = append(args, "--", "sh")
	log.Debug().Msgf("Shell args %v", args)
	runK(true, v.app, args...)
}

func (v *containerView) extraActions(aa keyActions) {
	aa[KeyL] = newKeyAction("Logs", v.logsCmd, true)
	aa[KeyShiftL] = newKeyAction("Previous Logs", v.prevLogsCmd, true)
	aa[KeyS] = newKeyAction("Shell", v.shellCmd, true)
	aa[tcell.KeyEscape] = newKeyAction("Back", v.backCmd, false)
	aa[KeyP] = newKeyAction("Previous", v.backCmd, false)
	aa[tcell.KeyEnter] = newKeyAction("View Logs", v.logsCmd, false)
	aa[KeyShiftC] = newKeyAction("Sort CPU", v.sortColCmd(7, false), true)
	aa[KeyShiftM] = newKeyAction("Sort MEM", v.sortColCmd(8, false), true)
}

func (v *containerView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.getTV()
		t.sortCol.index, t.sortCol.asc = t.nameColIndex()+col, asc
		t.refresh()

		return nil
	}
}

func (v *containerView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.app.inject(v.current)

	return nil
}
