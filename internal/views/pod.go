package views

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/watch"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const containerFmt = "[fg:bg:b]%s([hilite:bg:b]%s[fg:bg:-])"

type podView struct {
	*resourceView

	childCancelFn context.CancelFunc
}

var _ updatable = &podView{}

type loggable interface {
	appView() *appView
	backFn() actionHandler
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
	v.AddPage("logs", newLogsView(list.GetName(), &v), true, false)

	return &v
}

func (v *podView) extraActions(aa keyActions) {
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
	if !v.rowSelected() {
		return
	}

	po, err := v.app.informer.Get(watch.PodIndex, sel, metav1.GetOptions{})
	if err != nil {
		log.Error().Err(err).Msgf("Unable to retrieve pod %s", sel)
		app.flash().errf("Unable to retrieve pods %s", err)
		return
	}
	pod := po.(*v1.Pod)
	mx := k8s.NewMetricsServer(app.conn())
	list := resource.NewContainerList(app.conn(), mx, pod)
	title := skinTitle(fmt.Sprintf(containerFmt, "Containers", sel), v.app.styles.Style)

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
	v.switchPage("po")
	v.restartUpdates()
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
		v.app.flash().errf("Unable to retrieve containers %s", err)
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
		v.app.flash().errf("Unable to retrieve containers %s", err)
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
	v.stopUpdates()
	shellIn(v.app, path, co)
	v.restartUpdates()
}

func (v *podView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.getTV()
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

func computeShellArgs(path, co, context string, cfg *string) []string {
	a := make([]string, 0, 15)
	a = append(a, "exec", "-it")
	a = append(a, "--context", context)
	ns, po := namespaced(path)
	a = append(a, "-n", ns)
	a = append(a, po)
	if cfg != nil && *cfg != "" {
		a = append(a, "--kubeconfig", *cfg)
	}
	if co != "" {
		a = append(a, "-c", co)
	}
	a = append(a, "--", "sh")

	return a
}
