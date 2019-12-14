package view

import (
	"context"
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/watch"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	podTitle   = "Pods"
	shellCheck = "command -v bash >/dev/null && exec bash || exec sh"
)

// Pod represents a pod viewer.
type Pod struct {
	ResourceViewer
}

// NewPod returns a new viewer.
func NewPod(gvr dao.GVR) ResourceViewer {
	p := Pod{ResourceViewer: NewLogsExtender(NewBrowser(gvr), nil)}
	p.SetBindKeysFn(p.bindKeys)
	p.GetTable().SetEnterFn(p.showContainers)
	p.GetTable().SetColorerFn(render.Pod{}.ColorerFunc())

	return &p
}

func (p *Pod) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		tcell.KeyCtrlK: ui.NewKeyAction("Kill", p.killCmd, true),
		ui.KeyS:        ui.NewKeyAction("Shell", p.shellCmd, true),
		ui.KeyShiftR:   ui.NewKeyAction("Sort Ready", p.GetTable().SortColCmd(1, true), false),
		ui.KeyShiftS:   ui.NewKeyAction("Sort Status", p.GetTable().SortColCmd(2, true), false),
		ui.KeyShiftT:   ui.NewKeyAction("Sort Restart", p.GetTable().SortColCmd(3, false), false),
		ui.KeyShiftC:   ui.NewKeyAction("Sort CPU", p.GetTable().SortColCmd(4, false), false),
		ui.KeyShiftM:   ui.NewKeyAction("Sort MEM", p.GetTable().SortColCmd(5, false), false),
		ui.KeyShiftX:   ui.NewKeyAction("Sort CPU%", p.GetTable().SortColCmd(6, false), false),
		ui.KeyShiftZ:   ui.NewKeyAction("Sort MEM%", p.GetTable().SortColCmd(7, false), false),
		ui.KeyShiftI:   ui.NewKeyAction("Sort IP", p.GetTable().SortColCmd(8, true), false),
		ui.KeyShiftO:   ui.NewKeyAction("Sort Node", p.GetTable().SortColCmd(9, true), false),
	})
}

func (p *Pod) showContainers(app *App, ns, gvr, path string) {
	log.Debug().Msgf("SHOW CONTAINERS %q -- %q -- %q", gvr, ns, path)
	co := NewContainer(dao.GVR("containers"))
	co.SetContextFn(p.podContext)
	app.inject(co)
}

func (p *Pod) podContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeyPath, p.GetTable().GetSelectedItem())
}

// Commands...

func (p *Pod) killCmd(evt *tcell.EventKey) *tcell.EventKey {
	sels := p.GetTable().GetSelectedItems()
	if len(sels) == 0 {
		return evt
	}

	res, err := dao.AccessorFor(p.App().factory, dao.GVR(p.GVR()))
	if err != nil {
		p.App().Flash().Err(err)
		return nil
	}
	nuker, ok := res.(dao.Nuker)
	if !ok {
		p.App().Flash().Err(fmt.Errorf("expecting a nuker for %q", p.GVR()))
		return nil
	}
	p.GetTable().ShowDeleted()
	for _, res := range sels {
		p.App().Flash().Infof("Delete resource %s -- %s", p.GVR(), res)
		if err := nuker.Delete(res, true, false); err != nil {
			p.App().Flash().Errf("Delete failed with %s", err)
		} else {
			p.App().factory.DeleteForwarder(res)
		}
	}
	p.Refresh()

	return nil
}

func (p *Pod) shellCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := p.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	row := p.GetTable().GetSelectedRowIndex()
	status := ui.TrimCell(p.GetTable().SelectTable, row, p.GetTable().NameColIndex()+2)
	if status != render.Running {
		p.App().Flash().Errf("%s is not in a running state", sel)
		return nil
	}
	cc, err := fetchContainers(p.App().factory, sel, false)
	if err != nil {
		p.App().Flash().Errf("Unable to retrieve containers %s", err)
		return evt
	}
	if len(cc) == 1 {
		p.shellIn(sel, "")
		return nil
	}
	picker := NewPicker()
	picker.populate(cc)
	picker.SetSelectedFunc(func(i int, t, d string, r rune) {
		p.shellIn(sel, t)
	})
	p.App().inject(picker)

	return evt
}

func (p *Pod) shellIn(path, co string) {
	p.Stop()
	shellIn(p.App(), path, co)
	p.Start()
}

// ----------------------------------------------------------------------------
// Helpers...

func fetchContainers(f *watch.Factory, path string, includeInit bool) ([]string, error) {
	o, err := f.Get("v1/pods", path, labels.Everything())
	if err != nil {
		return nil, err
	}

	var pod v1.Pod
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pod)
	if err != nil {
		return nil, err
	}

	nn := make([]string, 0, len(pod.Spec.Containers)+len(pod.Spec.InitContainers))
	for _, c := range pod.Spec.Containers {
		nn = append(nn, c.Name)
	}
	if includeInit {
		for _, c := range pod.Spec.InitContainers {
			nn = append(nn, c.Name)
		}
	}
	return nn, nil
}

func shellIn(a *App, path, co string) {
	args := computeShellArgs(path, co, a.Config.K9s.CurrentContext, a.Conn().Config().Flags().KubeConfig)
	log.Debug().Msgf("Shell args %v", args)
	if !runK(true, a, args...) {
		a.Flash().Err(errors.New("Shell exec failed"))
	}
}

func computeShellArgs(path, co, context string, kcfg *string) []string {
	args := make([]string, 0, 15)
	args = append(args, "exec", "-it")
	args = append(args, "--context", context)
	ns, po := k8s.Namespaced(path)
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
