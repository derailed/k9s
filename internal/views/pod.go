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

func newPodView(title, gvr string, app *appView, list resource.List) resourceViewer {
	v := podView{resourceView: newResourceView(title, gvr, app, list).(*resourceView)}
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
	aa[tcell.KeyCtrlK] = ui.NewKeyAction("Kill", v.killCmd, true)
	aa[ui.KeyS] = ui.NewKeyAction("Shell", v.shellCmd, true)

	aa[ui.KeyL] = ui.NewKeyAction("Logs", v.logsCmd, true)
	aa[ui.KeyShiftL] = ui.NewKeyAction("Logs Previous", v.prevLogsCmd, true)

	aa[ui.KeyShiftR] = ui.NewKeyAction("Sort Ready", v.sortColCmd(1, false), false)
	aa[ui.KeyShiftS] = ui.NewKeyAction("Sort Status", v.sortColCmd(2, true), false)
	aa[ui.KeyShiftT] = ui.NewKeyAction("Sort Restart", v.sortColCmd(3, false), false)
	aa[ui.KeyShiftC] = ui.NewKeyAction("Sort CPU", v.sortColCmd(4, false), false)
	aa[ui.KeyShiftM] = ui.NewKeyAction("Sort MEM", v.sortColCmd(5, false), false)
	aa[ui.KeyShiftX] = ui.NewKeyAction("Sort CPU%", v.sortColCmd(6, false), false)
	aa[ui.KeyShiftZ] = ui.NewKeyAction("Sort MEM%", v.sortColCmd(7, false), false)
	aa[ui.KeyShiftD] = ui.NewKeyAction("Sort IP", v.sortColCmd(8, true), false)
	aa[ui.KeyShiftO] = ui.NewKeyAction("Sort Node", v.sortColCmd(9, true), false)
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
	return v.masterPage().GetSelectedItem()
}

func (v *podView) killCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() {
		return evt
	}

	sel := v.masterPage().GetSelectedItem()
	v.masterPage().ShowDeleted()
	v.app.Flash().Infof("Delete resource %s %s", v.list.GetName(), sel)
	if err := v.list.Resource().Delete(sel, true, false); err != nil {
		v.app.Flash().Errf("Delete failed with %s", err)
	} else {
		deletePortForward(v.app.forwarders, sel)
		v.refresh()
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
	if !v.masterPage().RowSelected() {
		return false
	}
	v.showLogs(v.masterPage().GetSelectedItem(), "", v, prev)

	return true
}

func (v *podView) showLogs(path, co string, parent loggable, prev bool) {
	l := v.GetPrimitive("logs").(*logsView)
	l.reload(co, parent, prev)
	v.switchPage("logs")
}

func (v *podView) shellCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() {
		return evt
	}

	sel := v.masterPage().GetSelectedItem()
	cc, err := fetchContainers(v.list, sel, false)
	if err != nil {
		v.app.Flash().Errf("Unable to retrieve containers %s", err)
		return evt
	}
	if len(cc) == 1 {
		v.shellIn(sel, "")
		return nil
	}
	p := v.GetPrimitive("picker").(*selectList)
	p.populate(cc)
	p.SetSelectedFunc(func(i int, t, d string, r rune) {
		v.shellIn(sel, t)
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
		t.SetSortCol(t.NameColIndex()+col, 0, asc)
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
