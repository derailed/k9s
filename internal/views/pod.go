package views

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/resource"
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
	v.AddPage("logs", newLogsView(list.GetName(), app, &v), true, false)

	return &v
}

func (v *podView) extraActions(aa keyActions) {
	// aa[KeyAltS] = newKeyAction("Sniff", v.sniffCmd, true)

	aa[KeyL] = newKeyAction("Logs", v.logsCmd, true)
	aa[KeyShiftL] = newKeyAction("Logs Previous", v.prevLogsCmd, true)
	aa[KeyS] = newKeyAction("Shell", v.shellCmd, true)
	aa[KeyShiftR] = newKeyAction("Sort Ready", v.sortColCmd(1, false), true)
	aa[KeyShiftS] = newKeyAction("Sort Status", v.sortColCmd(2, true), true)
	aa[KeyShiftT] = newKeyAction("Sort Restart", v.sortColCmd(3, false), true)
	aa[KeyShiftC] = newKeyAction("Sort CPU", v.sortColCmd(4, false), true)
	aa[KeyShiftM] = newKeyAction("Sort MEM", v.sortColCmd(5, false), true)
	aa[KeyAltC] = newKeyAction("Sort CPU%", v.sortColCmd(6, false), true)
	aa[KeyAltM] = newKeyAction("Sort MEM%", v.sortColCmd(7, false), true)
	aa[KeyShiftD] = newKeyAction("Sort IP", v.sortColCmd(8, true), true)
	aa[KeyShiftO] = newKeyAction("Sort Node", v.sortColCmd(9, true), true)
}

func (v *podView) listContainers(app *appView, _, res, sel string) {
	po, err := v.app.informer.Get(watch.PodIndex, sel, metav1.GetOptions{})
	if err != nil {
		app.flash().errf("Unable to retrieve pods %s", err)
		return
	}

	pod := po.(*v1.Pod)
	list := resource.NewContainerList(app.conn(), pod)
	title := skinTitle(fmt.Sprintf(containerFmt, "Containers", sel), app.styles.Frame())

	// Stop my updater
	if v.cancelFn != nil {
		v.cancelFn()
	}

	// Span child view
	cv := newContainerView(title, app, list, fqn(pod.Namespace, pod.Name), v.exitFn)
	v.AddPage("containers", cv, true, true)
	var ctx context.Context
	ctx, v.childCancelFn = context.WithCancel(v.parentCtx)
	cv.init(ctx, pod.Namespace)
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
		v.app.flash().info("Sniff launched!")
	} else {
		v.app.flash().info("Sniff failed!")
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

	r := v.selectedRow
	col := 2
	if v.list.AllNamespaces() {
		col = 3
	}
	status := strings.TrimSpace(v.masterPage().GetCell(r, col).Text)
	if status == "Running" || status == "Completed" {
		v.showLogs(v.selectedItem, "", v, prev)
		return true
	}

	v.app.flash().err(errors.New("Selected pod is not running"))
	return false
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
		v.app.flash().errf("Unable to retrieve containers %s", err)
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
		t.sortCol.index, t.sortCol.asc = t.nameColIndex()+col, asc
		t.refresh()

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
	args := computeShellArgs(path, co, a.config.K9s.CurrentContext, a.conn().Config().Flags().KubeConfig)
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
