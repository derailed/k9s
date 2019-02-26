package views

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	log "github.com/sirupsen/logrus"
)

type podView struct {
	*resourceView
}

type loggable interface {
	appView() *appView
	backFn() actionHandler
	getSelection() string
	getList() resource.List
	switchPage(n string)
}

func newPodView(t string, app *appView, list resource.List, c colorerFn) resourceViewer {
	v := podView{newResourceView(t, app, list, c).(*resourceView)}
	{
		v.extraActionsFn = v.extraActions
	}

	v.AddPage("logs", newLogsView(&v), true, false)

	picker := newSelectList()
	{
		picker.SetSelectedFunc(func(i int, t, d string, r rune) {
			v.sshInto(v.selectedItem, t)
		})
		picker.setActions(keyActions{
			tcell.KeyEscape: {description: "Back", action: v.backCmd},
		})
		v.AddPage("choose", picker, true, false)
	}

	v.switchPage("po")
	return &v
}

// Protocol...

func (v *podView) backFn() actionHandler {
	return v.backCmd
}

func (v *podView) appView() *appView {
	return v.app
}

func (v *podView) getList() resource.List {
	return v.list
}

func (v *podView) getSelection() string {
	return v.selectedItem
}

// Handlers...

func (v *podView) logsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	cc, err := fetchContainers(v.list, v.selectedItem, true)
	if err != nil {
		v.app.flash(flashErr, err.Error())
		log.Error(err)
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

func (v *podView) shellCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	cc, err := fetchContainers(v.list, v.selectedItem, false)
	if err != nil {
		v.app.flash(flashErr, err.Error())
		log.Error("Error fetching containers", err)
		return evt
	}
	if len(cc) == 1 {
		v.sshInto(v.selectedItem, "")
		return nil
	}
	v.showPicker(cc)
	return nil
}

func (v *podView) showPicker(cc []string) {
	l := v.GetPrimitive("choose").(*selectList)
	l.populate(cc)
	v.switchPage("choose")
}

func (v *podView) sshInto(path, co string) {
	ns, po := namespaced(path)
	if len(co) == 0 {
		run(v.app, "exec", "-it", "-n", ns, po, "--", "sh")
	} else {
		run(v.app, "exec", "-it", "-n", ns, po, "-c", co, "--", "sh")
	}
}

func (v *podView) extraActions(aa keyActions) {
	aa[KeyL] = newKeyAction("Logs", v.logsCmd)
	aa[KeyS] = newKeyAction("Shell", v.shellCmd)
}

func fetchContainers(l resource.List, po string, includeInit bool) ([]string, error) {
	if len(po) == 0 {
		return []string{}, nil
	}
	return l.Resource().(resource.Container).Containers(po, includeInit)
}
