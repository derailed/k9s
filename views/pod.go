package views

import (
	"github.com/derailed/k9s/resource"
	"github.com/gdamore/tcell"
	log "github.com/sirupsen/logrus"
)

type podView struct {
	*resourceView
}

func newPodView(t string, app *appView, list resource.List, c colorerFn) resourceViewer {
	v := podView{newResourceView(t, app, list, c).(*resourceView)}
	v.extraActionsFn = v.extraActions

	logs := newLogsView(&v)
	{
		logs.setActions(keyActions{
			tcell.KeyCtrlB: {description: "Back", action: v.stopLogs},
			tcell.KeyCtrlK: {description: "Clear", action: v.clearLogs},
		})
		v.AddPage("logs", logs, true, false)
	}

	picker := newSelectList()
	{
		picker.SetSelectedFunc(func(i int, t, d string, r rune) {
			log.Println("Selected", i, t, d, r)
			v.sshInto(v.selectedItem, t)
		})
		picker.setActions(keyActions{
			tcell.KeyCtrlB: {description: "Back", action: v.back},
		})
		v.AddPage("choose", picker, true, false)
	}

	v.switchPage("po")
	return &v
}

// Handlers...

func (v *podView) back(*tcell.EventKey) {
	v.switchPage(v.list.GetName())
}

func (v *podView) logs(*tcell.EventKey) {
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

func (v *podView) ssh(*tcell.EventKey) {
	if !v.rowSelected() {
		return
	}

	cc, err := fetchContainers(v.list, v.selectedItem)
	if err != nil {
		log.Error("Error fetching containers", err)
		return
	}

	if len(cc) == 1 {
		v.sshInto(v.selectedItem, "")
		return
	}

	v.showPicker(cc)
	return
}

func (v *podView) showPicker(cc []string) {
	l := v.GetPrimitive("choose").(*selectList)
	l.populate(cc)
	v.switchPage("choose")
}

func (v *podView) sshInto(path, co string) {
	v.app.flash(flashInfo, "SSH into pod", path)
	ns, po := namespaced(path)
	if len(co) == 0 {
		run(v.app, "exec", "-it", "-n", ns, po, "--", "sh")
	} else {
		run(v.app, "exec", "-it", "-n", ns, po, "-c", co, "--", "sh")
	}
}

func (v *podView) clearLogs(*tcell.EventKey) {
	v.app.flash(flashInfo, "Clearing logs...")
	v.GetPrimitive("logs").(*logsView).clearLogs()
}

func (v *podView) stopLogs(*tcell.EventKey) {
	v.GetPrimitive("logs").(*logsView).stop()
	v.switchPage(v.list.GetName())
}

func (v *podView) extraActions(aa keyActions) {
	aa[tcell.KeyCtrlL] = newKeyHandler("Logs", v.logs)
	aa[tcell.KeyCtrlS] = newKeyHandler("SSH", v.ssh)
}

func fetchContainers(l resource.List, po string) ([]string, error) {
	if len(po) == 0 {
		return []string{}, nil
	}
	return l.Resource().(resource.Container).Containers(po)
}
