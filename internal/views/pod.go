package views

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
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

func (v *podView) shellCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	cc, err := fetchContainers(v.list, v.selectedItem, false)
	if err != nil {
		v.app.flash(flashErr, err.Error())
		log.Error().Msgf("Error fetching containers %v", err)
		return evt
	}
	if len(cc) == 1 {
		v.shellIn(v.selectedItem, "")
	} else {
		p := v.GetPrimitive("choose").(*selectList)
		p.populate(cc)
		p.SetSelectedFunc(func(i int, t, d string, r rune) {
			v.shellIn(v.selectedItem, t)
		})
		v.switchPage("choose")
	}
	return evt
}

func (v *podView) showPicker(cc []string) {
	l := v.GetPrimitive("choose").(*selectList)
	l.populate(cc)
	v.switchPage("choose")
}

func (v *podView) shellIn(path, co string) {
	ns, po := namespaced(path)
	args := make([]string, 0, 12)
	args = append(args, "exec", "-it")
	args = append(args, "--context", config.Root.K9s.CurrentContext)
	args = append(args, "-n", ns)
	args = append(args, po)
	if len(co) != 0 {
		args = append(args, "-c", co)
	}
	args = append(args, "--", "sh")
	log.Debug().Msgf("Shell args %v", args)
	runK(v.app, args...)
}

func (v *podView) showLogs(path, co string, previous bool) {
	ns, po := namespaced(path)
	args := make([]string, 0, 10)
	args = append(args, "logs", "-f")
	args = append(args, "-n", ns)
	args = append(args, "--context", config.Root.K9s.CurrentContext)
	if len(co) != 0 {
		args = append(args, "-c", co)
		v.app.flash(flashInfo, fmt.Sprintf("Viewing logs from container %s on pod %s", co, po))
	} else {
		v.app.flash(flashInfo, fmt.Sprintf("Viewing logs from pod %s", po))
	}
	args = append(args, po)
	runK(v.app, args...)
}

func (v *podView) extraActions(aa keyActions) {
	aa[KeyL] = newKeyAction("Logs", v.logsCmd, true)
	aa[KeyS] = newKeyAction("Shell", v.shellCmd, true)
}

func fetchContainers(l resource.List, po string, includeInit bool) ([]string, error) {
	if len(po) == 0 {
		return []string{}, nil
	}
	return l.Resource().(resource.Container).Containers(po, includeInit)
}
