// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	"github.com/fatih/color"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	windowsOS        = "windows"
	powerShell       = "powershell"
	osSelector       = "kubernetes.io/os"
	osBetaSelector   = "beta." + osSelector
	trUpload         = "Upload"
	trDownload       = "Download"
	pfIndicator      = "[orange::b]Ⓕ"
	defaultTxRetries = 999
	magicPrompt      = "Yes Please!"
)

// Pod represents a pod viewer.
type Pod struct {
	ResourceViewer
}

// NewPod returns a new viewer.
func NewPod(gvr client.GVR) ResourceViewer {
	var p Pod
	p.ResourceViewer = NewPortForwardExtender(
		NewOwnerExtender(
			NewVulnerabilityExtender(
				NewImageExtender(
					NewLogsExtender(NewBrowser(gvr), p.logOptions),
				),
			),
		),
	)
	p.AddBindKeysFn(p.bindKeys)
	p.GetTable().SetEnterFn(p.showContainers)
	p.GetTable().SetDecorateFn(p.portForwardIndicator)

	return &p
}

func (p *Pod) portForwardIndicator(data *model1.TableData) {
	ff := p.App().factory.Forwarders()

	defer decorateCpuMemHeaderRows(p.App(), data)
	idx, ok := data.IndexOfHeader("PF")
	if !ok {
		return
	}

	data.RowsRange(func(_ int, re model1.RowEvent) bool {
		if ff.IsPodForwarded(re.Row.ID) {
			re.Row.Fields[idx] = pfIndicator
		}
		return true
	})
}

func (p *Pod) bindDangerousKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		tcell.KeyCtrlK: ui.NewKeyActionWithOpts(
			"Kill",
			p.killCmd,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			}),
		ui.KeyS: ui.NewKeyActionWithOpts(
			"Shell",
			p.shellCmd,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			}),
		ui.KeyA: ui.NewKeyActionWithOpts(
			"Attach",
			p.attachCmd,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			}),
		ui.KeyT: ui.NewKeyActionWithOpts(
			"Transfer",
			p.transferCmd,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			}),
		ui.KeyZ: ui.NewKeyActionWithOpts(
			"Sanitize",
			p.sanitizeCmd,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			}),
	})
}

func (p *Pod) bindKeys(aa *ui.KeyActions) {
	if !p.App().Config.K9s.IsReadOnly() {
		p.bindDangerousKeys(aa)
	}

	aa.Bulk(ui.KeyMap{
		ui.KeyO:      ui.NewKeyAction("Show Node", p.showNode, true),
		ui.KeyShiftR: ui.NewKeyAction("Sort Ready", p.GetTable().SortColCmd(readyCol, true), false),
		ui.KeyShiftT: ui.NewKeyAction("Sort Restart", p.GetTable().SortColCmd("RESTARTS", false), false),
		ui.KeyShiftS: ui.NewKeyAction("Sort Status", p.GetTable().SortColCmd(statusCol, true), false),
		ui.KeyShiftI: ui.NewKeyAction("Sort IP", p.GetTable().SortColCmd("IP", true), false),
		ui.KeyShiftO: ui.NewKeyAction("Sort Node", p.GetTable().SortColCmd("NODE", true), false),
	})
	aa.Merge(resourceSorters(p.GetTable()))
}

func (p *Pod) logOptions(prev bool) (*dao.LogOptions, error) {
	path := p.GetTable().GetSelectedItem()
	if path == "" {
		return nil, errors.New("you must provide a selection")
	}

	pod, err := fetchPod(p.App().factory, path)
	if err != nil {
		return nil, err
	}

	return podLogOptions(p.App(), path, prev, pod.ObjectMeta, pod.Spec), nil
}

func (p *Pod) showContainers(app *App, _ ui.Tabular, _ client.GVR, _ string) {
	co := NewContainer(client.NewGVR("containers"))
	co.SetContextFn(p.coContext)
	if err := app.inject(co, false); err != nil {
		app.Flash().Err(err)
	}
}

func (p *Pod) coContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeyPath, p.GetTable().GetSelectedItem())
}

// Handlers...

func (p *Pod) showNode(evt *tcell.EventKey) *tcell.EventKey {
	path := p.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}
	pod, err := fetchPod(p.App().factory, path)
	if err != nil {
		p.App().Flash().Err(err)
		return nil
	}
	if pod.Spec.NodeName == "" {
		p.App().Flash().Err(errors.New("no node assigned"))
		return nil
	}
	no := NewNode(client.NewGVR("v1/nodes"))
	no.SetInstance(pod.Spec.NodeName)
	//no.SetContextFn(nodeContext(pod.Spec.NodeName))
	if err := p.App().inject(no, false); err != nil {
		p.App().Flash().Err(err)
	}

	return nil
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
	for _, path := range selections {
		if err := nuker.Delete(context.Background(), path, nil, dao.NowGrace); err != nil {
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

	if err := containerShellIn(p.App(), p, path, ""); err != nil {
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

func (p *Pod) sanitizeCmd(evt *tcell.EventKey) *tcell.EventKey {
	res, err := dao.AccessorFor(p.App().factory, p.GVR())
	if err != nil {
		p.App().Flash().Err(err)
		return nil
	}
	s, ok := res.(dao.Sanitizer)
	if !ok {
		p.App().Flash().Err(fmt.Errorf("expecting a sanitizer for %q", p.GVR()))
		return nil
	}

	msg := fmt.Sprintf("Sanitize deletes all pods in completed/error state\nPlease enter [orange::b]%s[-::-] to proceed.", magicPrompt)
	dialog.ShowConfirmAck(p.App().App, p.App().Content.Pages, magicPrompt, true, "Sanitize", msg, func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*p.App().Conn().Config().CallTimeout())
		defer cancel()
		total, err := s.Sanitize(ctx, p.GetTable().GetModel().GetNamespace())
		if err != nil {
			p.App().Flash().Err(err)
			return
		}
		p.App().Flash().Infof("Sanitized %d %s", total, p.GVR())
		p.Refresh()
	}, func() {})

	return nil
}

func (p *Pod) transferCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := p.GetTable().GetSelectedItem()
	if path == "" {
		return nil
	}

	ns, n := client.Namespaced(path)
	ack := func(args dialog.TransferArgs) bool {
		local := args.To
		if !args.Download {
			local = args.From
		}
		if _, err := os.Stat(local); !args.Download && errors.Is(err, fs.ErrNotExist) {
			p.App().Flash().Err(err)
			return false
		}

		opts := make([]string, 0, 10)
		opts = append(opts, "cp")
		opts = append(opts, strings.TrimSpace(args.From))
		opts = append(opts, strings.TrimSpace(args.To))
		opts = append(opts, fmt.Sprintf("--no-preserve=%t", args.NoPreserve))
		opts = append(opts, fmt.Sprintf("--retries=%d", args.Retries))
		if args.CO != "" {
			opts = append(opts, "-c="+args.CO)
		}
		opts = append(opts, fmt.Sprintf("--retries=%d", args.Retries))

		cliOpts := shellOpts{
			background: true,
			args:       opts,
		}
		op := trUpload
		if args.Download {
			op = trDownload
		}

		fqn := path + ":" + args.CO
		if err := runK(p.App(), cliOpts); err != nil {
			p.App().cowCmd(err.Error())
		} else {
			p.App().Flash().Infof("%s successful on %s!", op, fqn)
		}
		return true
	}

	pod, err := fetchPod(p.App().factory, path)
	if err != nil {
		p.App().Flash().Err(err)
		return nil
	}

	opts := dialog.TransferDialogOpts{
		Title:      "Transfer",
		Containers: fetchContainers(pod.ObjectMeta, pod.Spec, false),
		Message:    "Download Files",
		Pod:        fmt.Sprintf("%s/%s:", ns, n),
		Ack:        ack,
		Retries:    defaultTxRetries,
		Cancel:     func() {},
	}
	dialog.ShowUploads(p.App().Styles.Dialog(), p.App().Content.Pages, opts)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func containerShellIn(a *App, comp model.Component, path, co string) error {
	if co != "" {
		resumeShellIn(a, comp, path, co)
		return nil
	}

	pod, err := fetchPod(a.factory, path)
	if err != nil {
		return err
	}
	cc := fetchContainers(pod.ObjectMeta, pod.Spec, false)
	if len(cc) == 1 {
		resumeShellIn(a, comp, path, cc[0])
		return nil
	}
	picker := NewPicker()
	picker.populate(cc)
	picker.SetSelectedFunc(func(_ int, co, _ string, _ rune) {
		resumeShellIn(a, comp, path, co)
	})

	return a.inject(picker, false)
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
	err = runK(a, shellOpts{clear: true, banner: c.Sprintf(bannerFmt, fqn, co), args: args})
	if err != nil {
		a.Flash().Errf("Shell exec failed: %s", err)
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
	cc := fetchContainers(pod.ObjectMeta, pod.Spec, false)
	if len(cc) == 1 {
		resumeAttachIn(a, comp, path, cc[0])
		return nil
	}
	picker := NewPicker()
	picker.populate(cc)
	picker.SetSelectedFunc(func(_ int, co, _ string, _ rune) {
		resumeAttachIn(a, comp, path, co)
	})
	if err := a.inject(picker, false); err != nil {
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
	if err := runK(a, shellOpts{clear: true, banner: c.Sprintf(bannerFmt, path, co), args: args}); err != nil {
		a.Flash().Errf("Attach exec failed: %s", err)
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
	if ns != client.BlankNamespace {
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

func fetchContainers(meta metav1.ObjectMeta, spec v1.PodSpec, allContainers bool) []string {
	nn := make([]string, 0, len(spec.Containers)+len(spec.InitContainers))

	// put the default container as the first entry
	defaultContainer, hasDefaultContainer := dao.GetDefaultContainer(meta, spec)
	if hasDefaultContainer {
		nn = append(nn, defaultContainer)
	}

	for _, c := range spec.Containers {
		if !hasDefaultContainer || c.Name != defaultContainer {
			nn = append(nn, c.Name)
		}
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
	if podOS, ok := osFromSelector(po.Spec.NodeSelector); ok {
		return podOS, nil
	}

	node, err := dao.FetchNode(context.Background(), f, po.Spec.NodeName)
	if err == nil {
		if nodeOS, ok := osFromSelector(node.Labels); ok {
			return nodeOS, nil
		}
	}

	return "", errors.New("no os information available")
}

func osFromSelector(s map[string]string) (string, bool) {
	if os, ok := s[osBetaSelector]; ok {
		return os, ok
	}

	os, ok := s[osSelector]
	return os, ok
}

func resourceSorters(t *Table) *ui.KeyActions {
	return ui.NewKeyActionsFromMap(ui.KeyMap{
		ui.KeyShiftC:   ui.NewKeyAction("Sort CPU", t.SortColCmd(cpuCol, false), false),
		ui.KeyShiftM:   ui.NewKeyAction("Sort MEM", t.SortColCmd(memCol, false), false),
		ui.KeyShiftX:   ui.NewKeyAction("Sort CPU/R", t.SortColCmd("%CPU/R", false), false),
		ui.KeyShiftZ:   ui.NewKeyAction("Sort MEM/R", t.SortColCmd("%MEM/R", false), false),
		tcell.KeyCtrlX: ui.NewKeyAction("Sort CPU/L", t.SortColCmd("%CPU/L", false), false),
		tcell.KeyCtrlQ: ui.NewKeyAction("Sort MEM/L", t.SortColCmd("%MEM/L", false), false),
	})
}
