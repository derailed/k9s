package views

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	log "github.com/sirupsen/logrus"
)

type podView struct {
	*resourceView
}

func newPodView(t string, app *appView, list resource.List, c colorerFn) resourceViewer {
	v := podView{newResourceView(t, app, list, c).(*resourceView)}
	v.extraActionsFn = v.extraActions

	v.AddPage("logs", newLogsView(&v), true, false)

	picker := newSelectList()
	{
		picker.SetSelectedFunc(func(i int, t, d string, r rune) {
			v.sshInto(v.selectedItem, t)
		})
		picker.setActions(keyActions{
			tcell.KeyEscape: {description: "Back", action: v.back},
		})
		v.AddPage("choose", picker, true, false)
	}

	v.switchPage("po")
	return &v
}

// Handlers...

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

func (v *podView) shell(*tcell.EventKey) {
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
	v.app.flash(flashInfo, "Shell into pod", path)
	ns, po := namespaced(path)
	if len(co) == 0 {
		run(v.app, "exec", "-it", "-n", ns, po, "--", "sh")
	} else {
		run(v.app, "exec", "-it", "-n", ns, po, "-c", co, "--", "sh")
	}
}

func (v *podView) extraActions(aa keyActions) {
	aa[KeyL] = newKeyHandler("Logs", v.logs)
	aa[KeyS] = newKeyHandler("Shell", v.shell)
}

func fetchContainers(l resource.List, po string) ([]string, error) {
	if len(po) == 0 {
		return []string{}, nil
	}
	return l.Resource().(resource.Container).Containers(po)
}
