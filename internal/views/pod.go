package views

import (
	"fmt"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const fmat = "[aqua::b]%s([fuchsia::b]%s[aqua::-])"

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

func newPodView(t string, app *appView, list resource.List) resourceViewer {
	v := podView{newResourceView(t, app, list).(*resourceView)}
	{
		v.extraActionsFn = v.extraActions
		v.enterFn = v.listContainers
	}

	picker := newSelectList(&v)
	{
		picker.setActions(keyActions{
			tcell.KeyEscape: {description: "Back", action: v.backCmd, visible: true},
		})
	}
	v.AddPage("picker", picker, true, false)

	v.AddPage("logs", newLogsView(list.GetName(), &v), true, false)
	v.switchPage("po")

	return &v
}

func (v *podView) listContainers(app *appView, _, res, sel string) {
	if !v.rowSelected() {
		return
	}

	log.Debug().Msgf("Selected %s", sel)
	ns, n := namespaced(sel)
	po, err := app.conn().DialOrDie().CoreV1().Pods(ns).Get(n, metav1.GetOptions{})
	if err != nil {
		log.Error().Err(err).Msgf("Unable to retrieve pod %s", sel)
		app.flash(flashErr, err.Error())
		return
	}

	mx := k8s.NewMetricsServer(app.conn())
	list := resource.NewContainerList(app.conn(), mx, po)
	app.inject(newContainerView(fmt.Sprintf(fmat, "Containers", sel), app, list, namespacedName(po.Namespace, po.Name)))
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
	if v.viewLogs(false) {
		return nil
	}
	return evt
}

func (v *podView) prevLogsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.viewLogs(true) {
		return nil
	}
	return evt
}

func (v *podView) viewLogs(prev bool) bool {
	if !v.rowSelected() {
		return false
	}

	cc, err := fetchContainers(v.list, v.selectedItem, true)
	if err != nil {
		v.app.flash(flashErr, err.Error())
		log.Error().Err(err)
		return false
	}

	if len(cc) == 1 {
		v.showLogs(v.selectedItem, cc[0], v.list.GetName(), v, prev)
		return true
	}

	picker := v.GetPrimitive("picker").(*selectList)
	picker.populate(cc)
	picker.SetSelectedFunc(func(i int, t, d string, r rune) {
		v.showLogs(v.selectedItem, t, "picker", picker, prev)
	})
	v.switchPage("picker")

	return true
}

func (v *podView) showLogs(path, co, view string, parent loggable, prev bool) {
	l := v.GetPrimitive("logs").(*logsView)
	l.reload(co, parent, view, prev)

	v.switchPage("logs")
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
		return nil
	}

	p := v.GetPrimitive("picker").(*selectList)
	p.populate(cc)
	p.SetSelectedFunc(func(i int, t, d string, r rune) {
		v.shellIn(v.selectedItem, t)
	})
	v.switchPage("picker")

	return evt
}

func (v *podView) shellIn(path, co string) {
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

func (v *podView) extraActions(aa keyActions) {
	aa[KeyL] = newKeyAction("Logs", v.logsCmd, true)
	aa[KeyShiftL] = newKeyAction("Logs", v.prevLogsCmd, true)
	aa[KeyS] = newKeyAction("Shell", v.shellCmd, true)
	aa[KeyShiftR] = newKeyAction("Sort Ready", v.sortColCmd(1, false), true)
	aa[KeyShiftS] = newKeyAction("Sort Status", v.sortColCmd(2, true), true)
	aa[KeyShiftT] = newKeyAction("Sort Restart", v.sortColCmd(3, false), true)
	aa[KeyShiftC] = newKeyAction("Sort CPU", v.sortColCmd(4, false), true)
	aa[KeyShiftM] = newKeyAction("Sort MEM", v.sortColCmd(5, false), true)
	aa[KeyShiftO] = newKeyAction("Sort Node", v.sortColCmd(7, true), true)
}

func (v *podView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.getTV()
		t.sortCol.index, t.sortCol.asc = t.nameColIndex()+col, asc
		t.refresh()

		return nil
	}
}

func fetchContainers(l resource.List, po string, includeInit bool) ([]string, error) {
	if len(po) == 0 {
		return []string{}, nil
	}
	return l.Resource().(resource.Containers).Containers(po, includeInit)
}
