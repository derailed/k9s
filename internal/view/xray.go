// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/k9s/internal/xray"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
	"github.com/sahilm/fuzzy"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const xrayTitle = "Xray"

var _ ResourceViewer = (*Xray)(nil)

// Xray represents an xray tree view.
type Xray struct {
	*ui.Tree

	app      *App
	gvr      client.GVR
	meta     metav1.APIResource
	model    *model.Tree
	cancelFn context.CancelFunc
	envFn    EnvFunc
}

// NewXray returns a new view.
func NewXray(gvr client.GVR) ResourceViewer {
	return &Xray{
		gvr:   gvr,
		Tree:  ui.NewTree(),
		model: model.NewTree(gvr),
	}
}

func (x *Xray) SetFilter(string)                 {}
func (x *Xray) SetLabelFilter(map[string]string) {}

// Init initializes the view.
func (x *Xray) Init(ctx context.Context) error {
	x.envFn = x.k9sEnv

	if err := x.Tree.Init(ctx); err != nil {
		return err
	}
	x.SetKeyListenerFn(x.keyEntered)

	var err error
	x.meta, err = dao.MetaAccess.MetaFor(x.gvr)
	if err != nil {
		return err
	}

	if x.app, err = extractApp(ctx); err != nil {
		return err
	}

	x.bindKeys()
	x.SetBackgroundColor(x.app.Styles.Xray().BgColor.Color())
	x.SetBorderColor(x.app.Styles.Xray().FgColor.Color())
	x.SetBorderFocusColor(x.app.Styles.Frame().Border.FocusColor.Color())
	x.SetGraphicsColor(x.app.Styles.Xray().GraphicColor.Color())
	x.SetTitle(fmt.Sprintf(" %s-%s ", xrayTitle, cases.Title(language.Und, cases.NoLower).String(x.gvr.R())))

	x.model.SetRefreshRate(time.Duration(x.app.Config.K9s.GetRefreshRate()) * time.Second)
	x.model.SetNamespace(client.CleanseNamespace(x.app.Config.ActiveNamespace()))
	x.model.AddListener(x)

	x.SetChangedFunc(func(n *tview.TreeNode) {
		spec, ok := n.GetReference().(xray.NodeSpec)
		if !ok {
			log.Error().Msgf("No ref found on node %s", n.GetText())
			return
		}
		x.SetSelectedItem(spec.AsPath())
		x.refreshActions()
	})
	x.refreshActions()

	return nil
}

// InCmdMode checks if prompt is active.
func (*Xray) InCmdMode() bool {
	return false
}

// ExtraHints returns additional hints.
func (x *Xray) ExtraHints() map[string]string {
	if x.app.Config.K9s.UI.NoIcons {
		return nil
	}
	return xray.EmojiInfo()
}

// SetInstance sets specific resource instance.
func (x *Xray) SetInstance(string) {}

func (x *Xray) bindKeys() {
	x.Actions().Bulk(ui.KeyMap{
		ui.KeySlash:     ui.NewSharedKeyAction("Filter Mode", x.activateCmd, false),
		tcell.KeyEscape: ui.NewSharedKeyAction("Filter Reset", x.resetCmd, false),
		tcell.KeyEnter:  ui.NewKeyAction("Goto", x.gotoCmd, true),
	})
}

func (x *Xray) keyEntered() {
	x.ClearSelection()
	x.update(x.filter(x.model.Peek()))
}

func (x *Xray) refreshActions() {
	aa := ui.NewKeyActions()

	defer func() {
		if err := pluginActions(x, aa); err != nil {
			log.Warn().Err(err).Msg("Plugins load failed")
		}
		if err := hotKeyActions(x, aa); err != nil {
			log.Warn().Err(err).Msg("HotKeys load failed")
		}

		x.Actions().Merge(aa)
		x.app.Menu().HydrateMenu(x.Hints())
	}()

	x.Actions().Clear()
	x.bindKeys()
	x.Tree.BindKeys()

	spec := x.selectedSpec()
	if spec == nil {
		return
	}

	gvr := spec.GVR()
	var err error
	x.meta, err = dao.MetaAccess.MetaFor(client.NewGVR(gvr))
	if err != nil {
		log.Warn().Msgf("NO meta for %q -- %s", gvr, err)
		return
	}

	if client.Can(x.meta.Verbs, "edit") {
		aa.Add(ui.KeyE, ui.NewKeyAction("Edit", x.editCmd, true))
	}
	if client.Can(x.meta.Verbs, "delete") {
		aa.Add(tcell.KeyCtrlD, ui.NewKeyAction("Delete", x.deleteCmd, true))
	}
	if !dao.IsK9sMeta(x.meta) {
		aa.Bulk(ui.KeyMap{
			ui.KeyY: ui.NewKeyAction(yamlAction, x.viewCmd, true),
			ui.KeyD: ui.NewKeyAction("Describe", x.describeCmd, true),
		})
	}

	switch gvr {
	case "v1/namespaces":
		x.Actions().Delete(tcell.KeyEnter)
	case "containers":
		x.Actions().Delete(tcell.KeyEnter)
		aa.Bulk(ui.KeyMap{
			ui.KeyS: ui.NewKeyAction("Shell", x.shellCmd, true),
			ui.KeyL: ui.NewKeyAction("Logs", x.logsCmd(false), true),
			ui.KeyP: ui.NewKeyAction("Logs Previous", x.logsCmd(true), true),
		})
	case "v1/pods":
		aa.Bulk(ui.KeyMap{
			ui.KeyS: ui.NewKeyAction("Shell", x.shellCmd, true),
			ui.KeyA: ui.NewKeyAction("Attach", x.attachCmd, true),
			ui.KeyL: ui.NewKeyAction("Logs", x.logsCmd(false), true),
			ui.KeyP: ui.NewKeyAction("Logs Previous", x.logsCmd(true), true),
		})
	}
	x.Actions().Merge(aa)
}

// GetSelectedPath returns the current selection as string.
func (x *Xray) GetSelectedPath() string {
	spec := x.selectedSpec()
	if spec == nil {
		return ""
	}
	return spec.Path()
}

func (x *Xray) selectedSpec() *xray.NodeSpec {
	node := x.GetCurrentNode()
	if node == nil {
		return nil
	}

	ref, ok := node.GetReference().(xray.NodeSpec)
	if !ok {
		log.Error().Msgf("Expecting a NodeSpec!")
		return nil
	}

	return &ref
}

// EnvFn returns an plugin env function if available.
func (x *Xray) EnvFn() EnvFunc {
	return x.envFn
}

func (x *Xray) k9sEnv() Env {
	env := k8sEnv(x.app.Conn().Config())

	spec := x.selectedSpec()
	if spec == nil {
		return env
	}

	env["FILTER"] = x.CmdBuff().GetText()
	if env["FILTER"] == "" {
		ns, n := client.Namespaced(spec.Path())
		env["NAMESPACE"], env["FILTER"] = ns, n
	}

	switch spec.GVR() {
	case "containers":
		_, co := client.Namespaced(spec.Path())
		env["CONTAINER"] = co
		ns, n := client.Namespaced(*spec.ParentPath())
		env["NAMESPACE"], env["POD"], env["NAME"] = ns, n, co
	default:
		ns, n := client.Namespaced(spec.Path())
		env["NAMESPACE"], env["NAME"] = ns, n
	}

	return env
}

// Aliases returns all available aliases.
func (x *Xray) Aliases() map[string]struct{} {
	return aliasesFor(x.meta, x.app.command.AliasesFor(x.meta.Name))
}

func (x *Xray) logsCmd(prev bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		spec := x.selectedSpec()
		if spec == nil {
			return nil
		}

		x.showLogs(spec, prev)

		return nil
	}
}

func (x *Xray) showLogs(spec *xray.NodeSpec, prev bool) {
	// Need to load and wait for pods
	path, co := spec.Path(), ""
	if spec.GVR() == "containers" {
		_, coName := client.Namespaced(spec.Path())
		path, co = *spec.ParentPath(), coName
	}

	ns, _ := client.Namespaced(path)
	_, err := x.app.factory.CanForResource(ns, "v1/pods", client.ListAccess)
	if err != nil {
		x.app.Flash().Err(err)
		return
	}

	opts := dao.LogOptions{
		Path:      path,
		Container: co,
		Previous:  prev,
	}
	if err := x.app.inject(NewLog(client.NewGVR("v1/pods"), &opts), false); err != nil {
		x.app.Flash().Err(err)
	}
}

func (x *Xray) shellCmd(evt *tcell.EventKey) *tcell.EventKey {
	spec := x.selectedSpec()
	if spec == nil {
		return nil
	}

	if spec.Status() != "ok" {
		x.app.Flash().Errf("%s is not in a running state", spec.Path())
		return nil
	}

	path, co := spec.Path(), ""
	if spec.GVR() == "containers" {
		_, co = client.Namespaced(spec.Path())
		path = *spec.ParentPath()
	}

	if err := containerShellIn(x.app, x, path, co); err != nil {
		x.app.Flash().Err(err)
	}

	return nil
}

func (x *Xray) attachCmd(evt *tcell.EventKey) *tcell.EventKey {
	spec := x.selectedSpec()
	if spec == nil {
		return nil
	}

	if spec.Status() != "ok" {
		x.app.Flash().Errf("%s is not in a running state", spec.Path())
		return nil
	}

	path, co := spec.Path(), ""
	if spec.GVR() == "containers" {
		path = *spec.ParentPath()
	}

	if err := containerAttachIn(x.app, x, path, co); err != nil {
		x.app.Flash().Err(err)
	}

	return nil
}

func (x *Xray) viewCmd(evt *tcell.EventKey) *tcell.EventKey {
	spec := x.selectedSpec()
	if spec == nil {
		return evt
	}

	ctx := x.defaultContext()
	raw, err := x.model.ToYAML(ctx, spec.GVR(), spec.Path())
	if err != nil {
		x.App().Flash().Errf("unable to get resource %q -- %s", spec.GVR(), err)
		return nil
	}

	details := NewDetails(x.app, yamlAction, spec.Path(), contentYAML, true).Update(raw)
	if err := x.app.inject(details, false); err != nil {
		x.app.Flash().Err(err)
	}

	return nil
}

func (x *Xray) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	spec := x.selectedSpec()
	if spec == nil {
		return evt
	}

	x.Stop()
	defer x.Start()
	{
		gvr := client.NewGVR(spec.GVR())
		meta, err := dao.MetaAccess.MetaFor(gvr)
		if err != nil {
			log.Warn().Msgf("NO meta for %q -- %s", spec.GVR(), err)
			return nil
		}
		x.resourceDelete(gvr, spec, fmt.Sprintf("Delete %s %s?", meta.SingularName, spec.Path()))
	}

	return nil
}

func (x *Xray) describeCmd(evt *tcell.EventKey) *tcell.EventKey {
	spec := x.selectedSpec()
	if spec == nil {
		return evt
	}

	x.describe(spec.GVR(), spec.Path())

	return nil
}

func (x *Xray) describe(gvr, path string) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, internal.KeyFactory, x.app.factory)

	yaml, err := x.model.Describe(ctx, gvr, path)
	if err != nil {
		x.app.Flash().Errf("Describe command failed: %s", err)
		return
	}

	details := NewDetails(x.app, "Describe", path, contentYAML, true).Update(yaml)
	if err := x.app.inject(details, false); err != nil {
		x.app.Flash().Err(err)
	}
}

func (x *Xray) editCmd(evt *tcell.EventKey) *tcell.EventKey {
	spec := x.selectedSpec()
	if spec == nil {
		return evt
	}

	x.Stop()
	defer x.Start()
	{
		ns, n := client.Namespaced(spec.Path())
		args := make([]string, 0, 10)
		args = append(args, "edit")
		args = append(args, client.NewGVR(spec.GVR()).R())
		args = append(args, "-n", ns)
		args = append(args, "--context", x.app.Config.K9s.ActiveContextName())
		if cfg := x.app.Conn().Config().Flags().KubeConfig; cfg != nil && *cfg != "" {
			args = append(args, "--kubeconfig", *cfg)
		}
		if err := runK(x.app, shellOpts{args: append(args, n)}); err != nil {
			x.app.Flash().Errf("Edit exec failed: %s", err)
		}
	}

	return evt
}

func (x *Xray) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if x.app.InCmdMode() {
		return evt
	}
	x.app.ResetPrompt(x.CmdBuff())

	return nil
}

func (x *Xray) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !x.CmdBuff().InCmdMode() {
		x.CmdBuff().Reset()
		return x.app.PrevCmd(evt)
	}
	x.CmdBuff().Reset()
	x.model.ClearFilter()
	x.Start()

	return nil
}

func (x *Xray) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	if x.CmdBuff().IsActive() {
		if internal.IsLabelSelector(x.CmdBuff().GetText()) {
			x.Start()
		}
		x.CmdBuff().SetActive(false)
		x.GetRoot().ExpandAll()

		return nil
	}

	spec := x.selectedSpec()
	if spec == nil {
		return nil
	}
	if len(strings.Split(spec.Path(), "/")) == 1 {
		return nil
	}
	x.app.gotoResource(client.NewGVR(spec.GVR()).R(), spec.Path(), false)

	return nil
}

func (x *Xray) filter(root *xray.TreeNode) *xray.TreeNode {
	q := x.CmdBuff().GetText()
	if x.CmdBuff().Empty() || internal.IsLabelSelector(q) {
		return root
	}

	x.UpdateTitle()
	if f, ok := internal.IsFuzzySelector(q); ok {
		return root.Filter(f, fuzzyFilter)
	}

	if internal.IsInverseSelector(q) {
		return root.Filter(q, rxInverseFilter)
	}

	return root.Filter(q, rxFilter)
}

// TreeNodeSelected callback for node selection.
func (x *Xray) TreeNodeSelected() {
	x.app.QueueUpdateDraw(func() {
		n := x.GetCurrentNode()
		if n != nil {
			n.SetColor(x.app.Styles.Xray().CursorColor.Color())
		}
	})
}

// TreeLoadFailed notifies the load failed.
func (x *Xray) TreeLoadFailed(err error) {
	x.app.Flash().Err(err)
}

func (x *Xray) update(node *xray.TreeNode) {
	root := makeTreeNode(node, x.ExpandNodes(), x.app.Config.K9s.UI.NoIcons, x.app.Styles)
	if node == nil {
		x.app.QueueUpdateDraw(func() {
			x.SetRoot(root)
		})
		return
	}

	for _, c := range node.Children {
		x.hydrate(root, c)
	}
	if x.GetSelectedItem() == "" {
		x.SetSelectedItem(node.Spec().Path())
	}

	x.app.QueueUpdateDraw(func() {
		x.SetRoot(root)
		root.Walk(func(node, parent *tview.TreeNode) bool {
			spec, ok := node.GetReference().(xray.NodeSpec)
			if !ok {
				log.Error().Msgf("Expecting a NodeSpec but got %T", node.GetReference())
				return false
			}
			// BOZO!! Figure this out expand/collapse but the root
			if parent != nil {
				node.SetExpanded(x.ExpandNodes())
			} else {
				node.SetExpanded(true)
			}

			if spec.AsPath() == x.GetSelectedItem() {
				node.SetExpanded(true).SetSelectable(true)
				x.SetCurrentNode(node)
			}
			return true
		})
	})
}

// TreeChanged notifies the model data changed.
func (x *Xray) TreeChanged(node *xray.TreeNode) {
	x.Count = node.Count(x.gvr.String())
	x.update(x.filter(node))
	x.UpdateTitle()
}

func (x *Xray) hydrate(parent *tview.TreeNode, n *xray.TreeNode) {
	node := makeTreeNode(n, x.ExpandNodes(), x.app.Config.K9s.UI.NoIcons, x.app.Styles)
	for _, c := range n.Children {
		x.hydrate(node, c)
	}
	parent.AddChild(node)
}

// SetEnvFn sets the custom environment function.
func (x *Xray) SetEnvFn(EnvFunc) {}

// Refresh updates the view.
func (x *Xray) Refresh() {}

// BufferCompleted indicates the buffer was changed.
func (x *Xray) BufferCompleted(_, _ string) {
	x.update(x.filter(x.model.Peek()))
}

// BufferChanged indicates the buffer was changed.
func (x *Xray) BufferChanged(_, _ string) {}

// BufferActive indicates the buff activity changed.
func (x *Xray) BufferActive(state bool, k model.BufferKind) {
	x.app.BufferActive(state, k)
}

func (x *Xray) defaultContext() context.Context {
	ctx := context.WithValue(context.Background(), internal.KeyFactory, x.app.factory)
	ctx = context.WithValue(ctx, internal.KeyFields, "")
	if x.CmdBuff().Empty() {
		ctx = context.WithValue(ctx, internal.KeyLabels, "")
	} else {
		ctx = context.WithValue(ctx, internal.KeyLabels, ui.TrimLabelSelector(x.CmdBuff().GetText()))
	}

	return ctx
}

// Start initializes resource watch loop.
func (x *Xray) Start() {
	x.Stop()
	x.CmdBuff().AddListener(x)

	ctx := x.defaultContext()
	ctx, x.cancelFn = context.WithCancel(ctx)
	x.model.Watch(ctx)
	x.UpdateTitle()
}

// Stop terminates watch loop.
func (x *Xray) Stop() {
	if x.cancelFn == nil {
		return
	}
	x.cancelFn()
	x.cancelFn = nil
	x.CmdBuff().RemoveListener(x)
}

// AddBindKeysFn sets up extra key bindings.
func (x *Xray) AddBindKeysFn(BindKeysFunc) {}

// SetContextFn sets custom context.
func (x *Xray) SetContextFn(ContextFunc) {}

// Name returns the component name.
func (x *Xray) Name() string { return "XRay" }

// GetTable returns the underlying table.
func (x *Xray) GetTable() *Table { return nil }

// GVR returns a resource descriptor.
func (x *Xray) GVR() client.GVR { return x.gvr }

// App returns the current app handle.
func (x *Xray) App() *App {
	return x.app
}

// UpdateTitle updates the view title.
func (x *Xray) UpdateTitle() {
	t := x.styleTitle()
	x.app.QueueUpdateDraw(func() {
		x.SetTitle(t)
	})
}

func (x *Xray) styleTitle() string {
	base := fmt.Sprintf("%s-%s", xrayTitle, cases.Title(language.Und, cases.NoLower).String(x.gvr.R()))
	ns := x.model.GetNamespace()
	if client.IsAllNamespaces(ns) {
		ns = client.NamespaceAll
	}

	var title string
	if ns == client.ClusterScope {
		title = ui.SkinTitle(fmt.Sprintf(ui.TitleFmt, base, render.AsThousands(int64(x.Count))), x.app.Styles.Frame())
	} else {
		title = ui.SkinTitle(fmt.Sprintf(ui.NSTitleFmt, base, ns, render.AsThousands(int64(x.Count))), x.app.Styles.Frame())
	}

	buff := x.CmdBuff().GetText()
	if buff == "" {
		return title
	}
	if internal.IsLabelSelector(buff) {
		buff = ui.TrimLabelSelector(buff)
	}

	return title + ui.SkinTitle(fmt.Sprintf(ui.SearchFmt, buff), x.app.Styles.Frame())
}

func (x *Xray) resourceDelete(gvr client.GVR, spec *xray.NodeSpec, msg string) {
	dialog.ShowDelete(x.app.Styles.Dialog(), x.app.Content.Pages, msg, func(propagation *metav1.DeletionPropagation, force bool) {
		x.app.Flash().Infof("Delete resource %s %s", spec.GVR(), spec.Path())
		accessor, err := dao.AccessorFor(x.app.factory, gvr)
		if err != nil {
			log.Error().Err(err).Msgf("No accessor")
			return
		}

		nuker, ok := accessor.(dao.Nuker)
		if !ok {
			x.app.Flash().Errf("Invalid nuker %T", accessor)
			return
		}
		grace := dao.DefaultGrace
		if force {
			grace = dao.ForceGrace
		}
		if err := nuker.Delete(context.Background(), spec.Path(), nil, grace); err != nil {
			x.app.Flash().Errf("Delete failed with `%s", err)
		} else {
			x.app.Flash().Infof("%s `%s deleted successfully", x.GVR(), spec.Path())
			x.app.factory.DeleteForwarder(spec.Path())
		}
		x.Refresh()
	}, func() {})
}

// ----------------------------------------------------------------------------
// Helpers...

func fuzzyFilter(q, path string) bool {
	q = strings.TrimSpace(q[2:])
	mm := fuzzy.Find(q, []string{path})

	return len(mm) > 0
}

func rxFilter(q, path string) bool {
	rx := regexp.MustCompile(`(?i)` + q)
	tokens := strings.Split(path, xray.PathSeparator)
	for _, t := range tokens {
		if rx.MatchString(t) {
			return true
		}
	}

	return false
}

func rxInverseFilter(q, path string) bool {
	q = strings.TrimSpace(q[1:])
	rx := regexp.MustCompile(`(?i)` + q)
	tokens := strings.Split(path, xray.PathSeparator)
	for _, t := range tokens {
		if rx.MatchString(t) {
			return false
		}
	}

	return true
}

func makeTreeNode(node *xray.TreeNode, expanded bool, showIcons bool, styles *config.Styles) *tview.TreeNode {
	n := tview.NewTreeNode("No data...")
	if node != nil {
		n.SetText(node.Title(showIcons))
		n.SetReference(node.Spec())
	}
	n.SetSelectable(true)
	n.SetExpanded(expanded)
	n.SetColor(styles.Xray().CursorColor.Color())
	n.SetSelectedFunc(func() {
		n.SetExpanded(!n.IsExpanded())
	})
	return n
}
