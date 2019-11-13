package view

import (
	"fmt"

	"github.com/derailed/k9s/internal/model"
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

type loggable interface {
	getSelection() string
	getList() resource.List
	Pop() (model.Component, bool)
}

// Pod represents a pod viewer.
type Pod struct {
	*Resource

	logs   *Logs
	picker *selectList
}

// NewPod returns a new viewer.
func NewPod(title, gvr string, list resource.List) ResourceViewer {
	p := Pod{
		Resource: NewResource(title, gvr, list),
	}
	p.extraActionsFn = p.extraActions
	p.enterFn = p.listContainers

	p.picker = newSelectList(&p)
	p.picker.setActions(ui.KeyActions{
		tcell.KeyEscape: {Description: "Back", Action: p.backCmd, Visible: true},
	})

	p.logs = NewLogs(list.GetName(), &p)

	return &p
}

func (p *Pod) extraActions(aa ui.KeyActions) {
	aa[tcell.KeyCtrlK] = ui.NewKeyAction("Kill", p.killCmd, true)
	aa[ui.KeyS] = ui.NewKeyAction("Shell", p.shellCmd, true)

	aa[ui.KeyL] = ui.NewKeyAction("Logs", p.logsCmd, true)
	aa[ui.KeyShiftL] = ui.NewKeyAction("Logs Previous", p.prevLogsCmd, true)

	aa[ui.KeyShiftR] = ui.NewKeyAction("Sort Ready", p.sortColCmd(1, false), false)
	aa[ui.KeyShiftS] = ui.NewKeyAction("Sort Status", p.sortColCmd(2, true), false)
	aa[ui.KeyShiftT] = ui.NewKeyAction("Sort Restart", p.sortColCmd(3, false), false)
	aa[ui.KeyShiftC] = ui.NewKeyAction("Sort CPU", p.sortColCmd(4, false), false)
	aa[ui.KeyShiftM] = ui.NewKeyAction("Sort MEM", p.sortColCmd(5, false), false)
	aa[ui.KeyShiftX] = ui.NewKeyAction("Sort CPU%", p.sortColCmd(6, false), false)
	aa[ui.KeyShiftZ] = ui.NewKeyAction("Sort MEM%", p.sortColCmd(7, false), false)
	aa[ui.KeyShiftD] = ui.NewKeyAction("Sort IP", p.sortColCmd(8, true), false)
	aa[ui.KeyShiftO] = ui.NewKeyAction("Sort Node", p.sortColCmd(9, true), false)
}

func (p *Pod) listContainers(app *App, _, res, sel string) {
	po, err := p.app.informer.Get(watch.PodIndex, sel, metav1.GetOptions{})
	if err != nil {
		app.Flash().Errf("Unable to retrieve pods %s", err)
		return
	}

	pod := po.(*v1.Pod)
	list := resource.NewContainerList(app.Conn(), pod)
	title := skinTitle(fmt.Sprintf(containerFmt, "Container", sel), app.Styles.Frame())

	// Stop my updater
	if p.cancelFn != nil {
		p.cancelFn()
	}

	// Span child view
	v := NewContainer(title, list, fqn(pod.Namespace, pod.Name))
	p.app.inject(v)
}

// Protocol...

func (p *Pod) getList() resource.List {
	return p.list
}

func (p *Pod) getSelection() string {
	return p.masterPage().GetSelectedItem()
}

func (p *Pod) killCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !p.masterPage().RowSelected() {
		return evt
	}
	sel := p.masterPage().GetSelectedItems()
	p.masterPage().ShowDeleted()
	for _, res := range sel {
		p.app.Flash().Infof("Delete resource %s %s", p.list.GetName(), res)
		if err := p.list.Resource().Delete(res, true, false); err != nil {
			p.app.Flash().Errf("Delete failed with %s", err)
		} else {
			deletePortForward(p.app.forwarders, res)
		}
	}
	p.refresh()

	return nil
}

func (p *Pod) logsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if p.viewLogs(false) {
		return nil
	}

	return evt
}

func (p *Pod) prevLogsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if p.viewLogs(true) {
		return nil
	}

	return evt
}

func (p *Pod) viewLogs(prev bool) bool {
	if !p.masterPage().RowSelected() {
		return false
	}
	p.showLogs(p.masterPage().GetSelectedItem(), "", p, prev)

	return true
}

func (p *Pod) showLogs(path, co string, parent loggable, prev bool) {
	l := p.GetPrimitive("logs").(*Logs)
	l.reload(co, parent, prev)
	p.Push(p.logs)
}

func (p *Pod) shellCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !p.masterPage().RowSelected() {
		return evt
	}

	sel := p.masterPage().GetSelectedItem()
	cc, err := fetchContainers(p.list, sel, false)
	if err != nil {
		p.app.Flash().Errf("Unable to retrieve containers %s", err)
		return evt
	}
	if len(cc) == 1 {
		p.shellIn(sel, "")
		return nil
	}
	picker := p.GetPrimitive("picker").(*selectList)
	picker.populate(cc)
	picker.SetSelectedFunc(func(i int, t, d string, r rune) {
		p.shellIn(sel, t)
	})
	p.Push(p.picker)

	return evt
}

func (p *Pod) shellIn(path, co string) {
	p.Stop()
	shellIn(p.app, path, co)
	p.Start()
}

func (p *Pod) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := p.masterPage()
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

func shellIn(a *App, path, co string) {
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
