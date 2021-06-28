package view

import (
	"context"
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/fatih/color"
	"github.com/gdamore/tcell/v2"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	windowsOS      = "windows"
	powerShell     = "powershell"
	osBetaSelector = "beta.kubernetes.io/os"
	osSelector     = "kubernetes.io/os"
)

// Pod represents a pod viewer.
type Pod struct {
	ResourceViewer
}

// NewPod returns a new viewer.
func NewPod(gvr client.GVR) ResourceViewer {
	var p Pod
	p.ResourceViewer = NewPortForwardExtender(
		NewImageExtender(
			NewLogsExtender(NewBrowser(gvr), p.logOptions),
		),
	)
	p.AddBindKeysFn(p.bindKeys)
	p.GetTable().SetEnterFn(p.showContainers)
	p.GetTable().SetColorerFn(render.Pod{}.ColorerFunc())
	p.GetTable().SetDecorateFn(p.portForwardIndicator)

	return &p
}

func (p *Pod) portForwardIndicator(data render.TableData) render.TableData {
	ff := p.App().factory.Forwarders()

	col := data.IndexOfHeader("PF")
	for _, re := range data.RowEvents {
		if ff.IsPodForwarded(re.Row.ID) {
			re.Row.Fields[col] = "[orange::b]Ⓕ"
		}
	}

	return decorateCpuMemHeaderRows(p.App(), data)
}

func (p *Pod) bindDangerousKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		tcell.KeyCtrlK: ui.NewKeyAction("Kill", p.killCmd, true),
		ui.KeyS:        ui.NewKeyAction("Shell", p.shellCmd, true),
		ui.KeyA:        ui.NewKeyAction("Attach", p.attachCmd, true),
	})
}

func (p *Pod) bindKeys(aa ui.KeyActions) {
	if !p.App().Config.K9s.IsReadOnly() {
		p.bindDangerousKeys(aa)
	}

	aa.Add(ui.KeyActions{
		ui.KeyF:      ui.NewKeyAction("Show PortForward", p.showPFCmd, true),
		ui.KeyShiftR: ui.NewKeyAction("Sort Ready", p.GetTable().SortColCmd(readyCol, true), false),
		ui.KeyShiftT: ui.NewKeyAction("Sort Restart", p.GetTable().SortColCmd("RESTARTS", false), false),
		ui.KeyShiftS: ui.NewKeyAction("Sort Status", p.GetTable().SortColCmd(statusCol, true), false),
		ui.KeyShiftI: ui.NewKeyAction("Sort IP", p.GetTable().SortColCmd("IP", true), false),
		ui.KeyShiftO: ui.NewKeyAction("Sort Node", p.GetTable().SortColCmd("NODE", true), false),
	})
	aa.Add(resourceSorters(p.GetTable()))
}

func (p *Pod) logOptions() (*dao.LogOptions, error) {
	path := p.GetTable().GetSelectedItem()
	if path == "" {
		return nil, errors.New("you must provide a selection")
	}

	pod, err := fetchPod(p.App().factory, path)
	if err != nil {
		return nil, err
	}

	cc := fetchContainers(pod.Spec, true)

	cfg := p.App().Config.K9s.Logger
	opts := dao.LogOptions{
		Path:            path,
		Lines:           int64(cfg.TailCount),
		SinceSeconds:    cfg.SinceSeconds,
		SingleContainer: len(cc) == 1,
		AllContainers:   false,
		ShowTimestamp:   cfg.ShowTime,
	}
	if c, ok := dao.GetDefaultLogContainer(pod.ObjectMeta, pod.Spec); ok {
		opts.Container, opts.DefaultContainer = c, c
	} else if len(cc) == 1 {
		opts.Container = cc[0]
	} else {
		opts.AllContainers = true
	}

	return &opts, nil
}

func (p *Pod) showContainers(app *App, model ui.Tabular, gvr, path string) {
	co := NewContainer(client.NewGVR("containers"))
	co.SetContextFn(p.coContext)
	if err := app.inject(co); err != nil {
		app.Flash().Err(err)
	}
}

func (p *Pod) coContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeyPath, p.GetTable().GetSelectedItem())
}

// Handlers...

func (p *Pod) showPFCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := p.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	if !p.App().factory.Forwarders().IsPodForwarded(path) {
		p.App().Flash().Errf("no portforwards defined")
		return nil
	}
	pf := NewPortForward(client.NewGVR("portforwards"))
	pf.SetContextFn(p.portForwardContext)
	if err := p.App().inject(pf); err != nil {
		p.App().Flash().Err(err)
	}

	return nil
}

func (p *Pod) portForwardContext(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, internal.KeyBenchCfg, p.App().BenchFile)
	return context.WithValue(ctx, internal.KeyPath, p.GetTable().GetSelectedItem())
}

func (p *Pod) killCmd(evt *tcell.EventKey) *tcell.EventKey {
	selections := p.GetTable().GetSelectedItems()
	if len(selections) == 0 {
		return evt
	}

	res, err := dao.AccessorFor(p.App().factory, p.GVR())
	if err != nil {
		p.App().Flash().Err(err)
		return nil
	}
	nuker, ok := res.(dao.Nuker)
	if !ok {
		p.App().Flash().Err(fmt.Errorf("expecting a nuker for %q", p.GVR()))
		return nil
	}
	if len(selections) > 1 {
		p.App().Flash().Infof("Delete %d marked %s", len(selections), p.GVR())
	} else {
		p.App().Flash().Infof("Delete resource %s %s", p.GVR(), selections[0])
	}
	p.GetTable().ShowDeleted()
	log.Debug().Msgf("SELS %v", selections)
	for _, path := range selections {
		if err := nuker.Delete(path, true, true); err != nil {
			p.App().Flash().Errf("Delete failed with %s", err)
		} else {
			p.App().factory.DeleteForwarder(path)
		}
		p.GetTable().DeleteMark(path)
	}
	p.Refresh()

	return nil
}

func (p *Pod) shellCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := p.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	if !podIsRunning(p.App().factory, path) {
		p.App().Flash().Errf("%s is not in a running state", path)
		return nil
	}

	if err := containerShellin(p.App(), p, path, ""); err != nil {
		p.App().Flash().Err(err)
	}

	return nil
}

func (p *Pod) attachCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := p.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	if !podIsRunning(p.App().factory, path) {
		p.App().Flash().Errf("%s is not in a happy state", path)
		return nil
	}

	if err := containerAttachIn(p.App(), p, path, ""); err != nil {
		p.App().Flash().Err(err)
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func containerShellin(a *App, comp model.Component, path, co string) error {
	if co != "" {
		resumeShellIn(a, comp, path, co)
		return nil
	}

	pod, err := fetchPod(a.factory, path)
	if err != nil {
		return err
	}
	cc := fetchContainers(pod.Spec, false)
	if len(cc) == 1 {
		resumeShellIn(a, comp, path, cc[0])
		return nil
	}
	picker := NewPicker()
	picker.populate(cc)
	picker.SetSelectedFunc(func(_ int, co, _ string, _ rune) {
		resumeShellIn(a, comp, path, co)
	})

	return a.inject(picker)
}

func resumeShellIn(a *App, c model.Component, path, co string) {
	c.Stop()
	defer c.Start()

	shellIn(a, path, co)
}

func shellIn(a *App, fqn, co string) {
	os, err := getPodOS(a.factory, fqn)
	if err != nil {
		log.Warn().Err(err).Msgf("os detect failed")
	}
	args := computeShellArgs(fqn, co, a.Conn().Config().Flags().KubeConfig, os)

	c := color.New(color.BgGreen).Add(color.FgBlack).Add(color.Bold)
	if !runK(a, shellOpts{clear: true, banner: c.Sprintf(bannerFmt, fqn, co), args: args}) {
		a.Flash().Err(errors.New("Shell exec failed"))
	}
}

func containerAttachIn(a *App, comp model.Component, path, co string) error {
	if co != "" {
		resumeAttachIn(a, comp, path, co)
		return nil
	}

	pod, err := fetchPod(a.factory, path)
	if err != nil {
		return err
	}
	cc := fetchContainers(pod.Spec, false)
	if len(cc) == 1 {
		resumeAttachIn(a, comp, path, cc[0])
		return nil
	}
	picker := NewPicker()
	picker.populate(cc)
	picker.SetSelectedFunc(func(_ int, co, _ string, _ rune) {
		resumeAttachIn(a, comp, path, co)
	})
	if err := a.inject(picker); err != nil {
		return err
	}

	return nil
}

func resumeAttachIn(a *App, c model.Component, path, co string) {
	c.Stop()
	defer c.Start()

	attachIn(a, path, co)
}

func attachIn(a *App, path, co string) {
	args := buildShellArgs("attach", path, co, a.Conn().Config().Flags().KubeConfig)
	c := color.New(color.BgGreen).Add(color.FgBlack).Add(color.Bold)
	if !runK(a, shellOpts{clear: true, banner: c.Sprintf(bannerFmt, path, co), args: args}) {
		a.Flash().Err(errors.New("Attach exec failed"))
	}
}

func computeShellArgs(path, co string, kcfg *string, os string) []string {
	args := buildShellArgs("exec", path, co, kcfg)
	if os == windowsOS {
		return append(args, "--", powerShell)
	}
	return append(args, "--", "sh", "-c", shellCheck)
}

func buildShellArgs(cmd, path, co string, kcfg *string) []string {
	args := make([]string, 0, 15)
	args = append(args, cmd, "-it")
	ns, po := client.Namespaced(path)
	if ns != client.AllNamespaces {
		args = append(args, "-n", ns)
	}
	args = append(args, po)
	if kcfg != nil && *kcfg != "" {
		args = append(args, "--kubeconfig", *kcfg)
	}
	if co != "" {
		args = append(args, "-c", co)
	}

	return args
}

func fetchContainers(spec v1.PodSpec, allContainers bool) []string {
	nn := make([]string, 0, len(spec.Containers)+len(spec.InitContainers))
	for _, c := range spec.Containers {
		nn = append(nn, c.Name)
	}
	if !allContainers {
		return nn
	}
	for _, c := range spec.InitContainers {
		nn = append(nn, c.Name)
	}
	for _, c := range spec.EphemeralContainers {
		nn = append(nn, c.Name)
	}

	return nn
}

func fetchPod(f dao.Factory, path string) (*v1.Pod, error) {
	o, err := f.Get("v1/pods", path, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var pod v1.Pod
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pod)
	if err != nil {
		return nil, err
	}

	return &pod, nil
}

func podIsRunning(f dao.Factory, path string) bool {
	po, err := fetchPod(f, path)
	if err != nil {
		log.Error().Err(err).Msg("unable to fetch pod")
		return false
	}

	var re render.Pod
	return re.Phase(po) == render.Running
}

func getPodOS(f dao.Factory, fqn string) (string, error) {
	po, err := fetchPod(f, fqn)
	if err != nil {
		return "", err
	}
	if os, ok := po.Spec.NodeSelector[osBetaSelector]; ok {
		return os, nil
	}
	os, ok := po.Spec.NodeSelector[osSelector]
	if !ok {
		return "", fmt.Errorf("no os information available")
	}

	return os, nil
}

func resourceSorters(t *Table) ui.KeyActions {
	return ui.KeyActions{
		ui.KeyShiftC:   ui.NewKeyAction("Sort CPU", t.SortColCmd(cpuCol, false), false),
		ui.KeyShiftM:   ui.NewKeyAction("Sort MEM", t.SortColCmd(memCol, false), false),
		ui.KeyShiftX:   ui.NewKeyAction("Sort CPU/R", t.SortColCmd("%CPU/R", false), false),
		ui.KeyShiftZ:   ui.NewKeyAction("Sort MEM/R", t.SortColCmd("%MEM/R", false), false),
		tcell.KeyCtrlX: ui.NewKeyAction("Sort CPU/L", t.SortColCmd("%CPU/L", false), false),
		tcell.KeyCtrlQ: ui.NewKeyAction("Sort MEM/L", t.SortColCmd("%MEM/L", false), false),
	}
}
