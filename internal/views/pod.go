package views

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/watch"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	containerFmt = "[fg:bg:b]%s([hilite:bg:b]%s[fg:bg:-])"
	shellCheck   = "command -v bash >/dev/null && exec bash || exec sh"
)

type podView struct {
	*resourceView

	childCancelFn context.CancelFunc
}

var _ updatable = &podView{}

type loggable interface {
	getSelection() string
	getList() resource.List
	switchPage(n string)
}

func newPodView(t string, app *appView, list resource.List) resourceViewer {
	v := podView{resourceView: newResourceView(t, app, list).(*resourceView)}
	v.extraActionsFn = v.extraActions
	v.enterFn = v.listContainers

	picker := newSelectList(&v)
	{
		picker.setActions(ui.KeyActions{
			tcell.KeyEscape: {Description: "Back", Action: v.backCmd, Visible: true},
		})
	}
	v.AddPage("picker", picker, true, false)
	v.AddPage("logs", newLogsView(list.GetName(), app, &v), true, false)

	return &v
}

func (v *podView) extraActions(aa ui.KeyActions) {
	actions := ui.KeyActions{
		ui.KeyL:      ui.NewKeyAction("Logs", v.logsCmd, true),
		ui.KeyShiftL: ui.NewKeyAction("Logs Previous", v.prevLogsCmd, true),
		ui.KeyS:      ui.NewKeyAction("Shell", v.shellCmd, true),
		ui.KeyShiftR: ui.NewKeyAction("Sort Ready", v.sortColCmd(1, false), true),
		ui.KeyShiftS: ui.NewKeyAction("Sort Status", v.sortColCmd(2, true), true),
		ui.KeyShiftT: ui.NewKeyAction("Sort Restart", v.sortColCmd(3, false), true),
		ui.KeyShiftC: ui.NewKeyAction("Sort CPU", v.sortColCmd(4, false), true),
		ui.KeyShiftM: ui.NewKeyAction("Sort MEM", v.sortColCmd(5, false), true),
		ui.KeyAltC:   ui.NewKeyAction("Sort CPU%", v.sortColCmd(6, false), true),
		ui.KeyAltM:   ui.NewKeyAction("Sort MEM%", v.sortColCmd(7, false), true),
		ui.KeyShiftD: ui.NewKeyAction("Sort IP", v.sortColCmd(8, true), true),
		ui.KeyShiftO: ui.NewKeyAction("Sort Node", v.sortColCmd(9, true), true),
	}
	for k, v := range actions {
		aa[k] = v
	}
}

func (v *podView) listContainers(app *appView, _, res, sel string) {
	po, err := v.app.informer.Get(watch.PodIndex, sel, metav1.GetOptions{})
	if err != nil {
		app.Flash().Errf("Unable to retrieve pods %s", err)
		return
	}

	pod := po.(*v1.Pod)
	list := resource.NewContainerList(app.Conn(), pod)
	title := skinTitle(fmt.Sprintf(containerFmt, "Containers", sel), app.Styles.Frame())

	// Stop my updater
	if v.cancelFn != nil {
		v.cancelFn()
	}

	// Span child view
	cv := newContainerView(title, app, list, fqn(pod.Namespace, pod.Name), v.exitFn)
	v.AddPage("containers", cv, true, true)
	var ctx context.Context
	ctx, v.childCancelFn = context.WithCancel(v.parentCtx)
	cv.Init(ctx, pod.Namespace)
}

func (v *podView) exitFn() {
	if v.childCancelFn != nil {
		v.childCancelFn()
	}
	v.RemovePage("containers")
	v.switchPage("master")
	v.restartUpdates()
}

// Protocol...

func (v *podView) getList() resource.List {
	return v.list
}

func (v *podView) getSelection() string {
	return v.selectedItem
}

func (v *podView) sniffCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	ns, n := namespaced(v.selectedItem)
	var args []string
	args = append(args, "sniff", n, "-n", ns)

	if runK(true, v.app, args...) {
		v.app.Flash().Info("Sniff launched!")
	} else {
		v.app.Flash().Info("Sniff failed!")
	}
	return nil
}

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
	v.showLogs(v.selectedItem, "", v, prev)

	return true
}

func (v *podView) showLogs(path, co string, parent loggable, prev bool) {
	l := v.GetPrimitive("logs").(*logsView)
	l.reload(co, parent, prev)
	v.switchPage("logs")
}

func (v *podView) shellCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	cc, err := fetchContainers(v.list, v.selectedItem, false)
	if err != nil {
		v.app.Flash().Errf("Unable to retrieve containers %s", err)
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
	v.stopUpdates()
	shellIn(v.app, path, co)
	v.restartUpdates()
}

func (v *podView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.masterPage()
		t.SetSortCol(t.NameColIndex()+col, asc)
		t.Refresh()

		return nil
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func fetchContainers(l resource.List, po string, includeInit bool) ([]string, error) {
	if len(po) == 0 {
		return []string{}, nil
	}
	return l.Resource().(resource.Containers).Containers(po, includeInit)
}

func shellIn(a *appView, path, co string) {
	args := computeShellArgs(path, co, a.Config.K9s.CurrentContext, a.Conn().Config().Flags().KubeConfig)
	log.Debug().Msgf("Shell args %v", args)
	runK(true, a, args...)
}

func computeShellArgs(path, co, context string, kcfg *string) []string {
	args := make([]string, 0, 15)
	args = append(args, "exec", "-it")
	args = append(args, "--context", context)
	ns, po := namespaced(path)
	args = append(args, "-n", ns)
	args = append(args, po)
	if kcfg != nil && *kcfg != "" {
		args = append(args, "--kubeconfig", *kcfg)
	}
	if co != "" {
		args = append(args, "-c", co)
	}

	return append(args, "--", "sh", "-c", shellCheck)
}
