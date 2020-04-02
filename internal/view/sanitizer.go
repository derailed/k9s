package view

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/xray"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ ResourceViewer = (*Sanitizer)(nil)

// Sanitizer represents an sanitizer tree view.
type Sanitizer struct {
	*ui.Tree

	app       *App
	gvr       client.GVR
	meta      metav1.APIResource
	model     *model.Tree
	cancelFn  context.CancelFunc
	envFn     EnvFunc
	contextFn ContextFunc
}

// NewSanitizer returns a new view.
func NewSanitizer(gvr client.GVR) ResourceViewer {
	return &Sanitizer{
		gvr:   gvr,
		Tree:  ui.NewTree(),
		model: model.NewTree(gvr),
	}
}

// Init initializes the view
func (s *Sanitizer) Init(ctx context.Context) error {
	s.envFn = s.k9sEnv

	if err := s.Tree.Init(ctx); err != nil {
		return err
	}
	s.SetKeyListenerFn(s.keyEntered)

	var err error
	s.meta, err = dao.MetaAccess.MetaFor(s.gvr)
	if err != nil {
		return err
	}

	if s.app, err = extractApp(ctx); err != nil {
		return err
	}

	s.bindKeys()
	s.SetBackgroundColor(s.app.Styles.Xray().BgColor.Color())
	s.SetBorderColor(s.app.Styles.Xray().FgColor.Color())
	s.SetBorderFocusColor(s.app.Styles.Frame().Border.FocusColor.Color())
	s.SetGraphicsColor(s.app.Styles.Xray().GraphicColor.Color())
	s.SetTitle(strings.Title(s.gvr.R()))

	s.model.SetNamespace(client.CleanseNamespace(s.app.Config.ActiveNamespace()))
	s.model.AddListener(s)

	s.SetChangedFunc(func(n *tview.TreeNode) {
		spec, ok := n.GetReference().(xray.NodeSpec)
		if !ok {
			log.Error().Msgf("No ref found on node %s", n.GetText())
			return
		}
		s.SetSelectedItem(spec.AsPath())
		s.refreshActions()
	})
	s.refreshActions()

	return nil
}

// ExtraHints returns additional hints.
func (s *Sanitizer) ExtraHints() map[string]string {
	if s.app.Config.K9s.NoIcons {
		return nil
	}
	return xray.EmojiInfo()
}

// SetInstance sets specific resource instance.
func (s *Sanitizer) SetInstance(string) {}

func (s *Sanitizer) bindKeys() {
	s.Actions().Add(ui.KeyActions{
		tcell.KeyEnter:      ui.NewKeyAction("Goto", s.gotoCmd, true),
		ui.KeySlash:         ui.NewSharedKeyAction("Filter Mode", s.activateCmd, false),
		tcell.KeyBackspace2: ui.NewSharedKeyAction("Erase", s.eraseCmd, false),
		tcell.KeyBackspace:  ui.NewSharedKeyAction("Erase", s.eraseCmd, false),
		tcell.KeyDelete:     ui.NewSharedKeyAction("Erase", s.eraseCmd, false),
		tcell.KeyCtrlU:      ui.NewSharedKeyAction("Clear Filter", s.clearCmd, false),
		tcell.KeyCtrlW:      ui.NewSharedKeyAction("Clear Filter", s.clearCmd, false),
		tcell.KeyEscape:     ui.NewSharedKeyAction("Filter Reset", s.resetCmd, false),
	})
}

func (s *Sanitizer) keyEntered() {
	s.ClearSelection()
	s.update(s.filter(s.model.Peek()))
}

func (s *Sanitizer) refreshActions() {
}

// GetSelectedPath returns the current selection as string.
func (s *Sanitizer) GetSelectedPath() string {
	spec := s.selectedSpec()
	if spec == nil {
		return ""
	}
	return spec.Path()
}

func (s *Sanitizer) selectedSpec() *xray.NodeSpec {
	node := s.GetCurrentNode()
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
func (s *Sanitizer) EnvFn() EnvFunc {
	return s.envFn
}

func (s *Sanitizer) k9sEnv() Env {
	env := k8sEnv(s.app.Conn().Config())

	spec := s.selectedSpec()
	if spec == nil {
		return env
	}

	env["FILTER"] = s.CmdBuff().String()
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
func (s *Sanitizer) Aliases() []string {
	return append(s.meta.ShortNames, s.meta.SingularName, s.meta.Name)
}

func (s *Sanitizer) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if s.app.InCmdMode() {
		return evt
	}
	s.CmdBuff().SetActive(true)

	return nil
}

func (s *Sanitizer) clearCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !s.CmdBuff().IsActive() {
		return evt
	}
	s.CmdBuff().Clear()
	s.model.ClearFilter()
	s.Start()

	return nil
}

func (s *Sanitizer) eraseCmd(evt *tcell.EventKey) *tcell.EventKey {
	if s.CmdBuff().IsActive() {
		s.CmdBuff().Delete()
	}
	s.UpdateTitle()

	return nil
}

func (s *Sanitizer) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !s.CmdBuff().InCmdMode() {
		s.CmdBuff().Reset()
		return s.app.PrevCmd(evt)
	}
	s.CmdBuff().Reset()
	s.model.ClearFilter()
	s.Start()

	return nil
}

func (s *Sanitizer) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	if s.CmdBuff().IsActive() {
		if ui.IsLabelSelector(s.CmdBuff().String()) {
			s.Start()
		}
		s.CmdBuff().SetActive(false)
		s.GetRoot().ExpandAll()
		return nil
	}

	spec := s.selectedSpec()
	if spec == nil {
		return nil
	}
	if len(spec.GVRs) <= 2 {
		return nil
	}
	path := strings.Replace(spec.Path(), "::", "/", 1)
	if strings.Contains(path, "[") {
		return nil
	}
	if len(strings.Split(path, "/")) == 1 && spec.GVR() != "node" {
		path = "-/" + path
	}
	if err := s.app.gotoResource(client.NewGVR(spec.GVR()).R(), path, false); err != nil {
		log.Debug().Err(err)
	}

	return nil
}

func (s *Sanitizer) filter(root *xray.TreeNode) *xray.TreeNode {
	q := s.CmdBuff().String()
	if s.CmdBuff().Empty() || ui.IsLabelSelector(q) {
		return root
	}

	s.UpdateTitle()
	if ui.IsFuzzySelector(q) {
		return root.Filter(q, fuzzyFilter)
	}

	return root.Filter(q, rxFilter)
}

// TreeNodeSelected callback for node selection.
func (s *Sanitizer) TreeNodeSelected() {
	s.app.QueueUpdateDraw(func() {
		n := s.GetCurrentNode()
		if n != nil {
			n.SetColor(s.app.Styles.Xray().CursorColor.Color())
		}
	})
}

// TreeLoadFailed notifies the load failed.
func (s *Sanitizer) TreeLoadFailed(err error) {
	s.app.Flash().Err(err)
}

func (s *Sanitizer) update(node *xray.TreeNode) {
	root := makeTreeNode(node, s.ExpandNodes(), s.app.Config.K9s.NoIcons, s.app.Styles)
	if node == nil {
		s.app.QueueUpdateDraw(func() {
			s.SetRoot(root)
		})
		return
	}

	for _, c := range node.Children {
		s.hydrate(root, c)
	}
	if s.GetSelectedItem() == "" {
		s.SetSelectedItem(node.Spec().Path())
	}

	s.app.QueueUpdateDraw(func() {
		s.SetRoot(root)
		root.Walk(func(node, parent *tview.TreeNode) bool {
			spec, ok := node.GetReference().(xray.NodeSpec)
			if !ok {
				log.Error().Msgf("Expecting a NodeSpec but got %T", node.GetReference())
				return false
			}
			// BOZO!! Figure this out expand/collapse but the root
			if parent != nil {
				node.SetExpanded(s.ExpandNodes())
			} else {
				node.SetExpanded(true)
			}

			if spec.AsPath() == s.GetSelectedItem() {
				node.SetExpanded(true).SetSelectable(true)
				s.SetCurrentNode(node)
			}
			return true
		})
	})
}

// TreeChanged notifies the model data changed.
func (s *Sanitizer) TreeChanged(node *xray.TreeNode) {
	s.Count = node.Count(s.gvr.String())
	s.update(s.filter(node))
	s.UpdateTitle()
}

func (s *Sanitizer) hydrate(parent *tview.TreeNode, n *xray.TreeNode) {
	node := makeTreeNode(n, s.ExpandNodes(), s.app.Config.K9s.NoIcons, s.app.Styles)
	for _, c := range n.Children {
		s.hydrate(node, c)
	}
	parent.AddChild(node)
}

// SetEnvFn sets the custom environment function.
func (s *Sanitizer) SetEnvFn(EnvFunc) {}

// Refresh updates the view
func (s *Sanitizer) Refresh() {
}

// BufferChanged indicates the buffer was changed.
func (s *Sanitizer) BufferChanged(t string) {}

// BufferActive indicates the buff activity changed.
func (s *Sanitizer) BufferActive(state bool, k model.BufferKind) {
	s.app.BufferActive(state, k)
}

func (s *Sanitizer) defaultContext() context.Context {
	ctx := context.WithValue(context.Background(), internal.KeyFactory, s.app.factory)
	ctx = context.WithValue(ctx, internal.KeyFields, "")
	if s.CmdBuff().Empty() {
		ctx = context.WithValue(ctx, internal.KeyLabels, "")
	} else {
		ctx = context.WithValue(ctx, internal.KeyLabels, ui.TrimLabelSelector(s.CmdBuff().String()))
	}

	return ctx
}

// Start initializes resource watch loop.
func (s *Sanitizer) Start() {
	s.Stop()

	s.CmdBuff().AddListener(s.app.Cmd())
	s.CmdBuff().AddListener(s)

	ctx := s.defaultContext()
	ctx, s.cancelFn = context.WithCancel(ctx)
	if s.contextFn != nil {
		ctx = s.contextFn(ctx)
	}
	s.model.Refresh(ctx)
	s.UpdateTitle()
}

// Stop terminates watch loop.
func (s *Sanitizer) Stop() {
	if s.cancelFn == nil {
		return
	}
	s.cancelFn()
	s.cancelFn = nil

	s.CmdBuff().RemoveListener(s.app.Cmd())
	s.CmdBuff().RemoveListener(s)
}

// SetBindKeysFn sets up extra key bindings.
func (s *Sanitizer) SetBindKeysFn(BindKeysFunc) {}

// SetContextFn sets custom context.
func (s *Sanitizer) SetContextFn(f ContextFunc) {
	s.contextFn = f
}

// Name returns the component name.
func (s *Sanitizer) Name() string { return "report" }

// GetTable returns the underlying table.
func (s *Sanitizer) GetTable() *Table { return nil }

// GVR returns a resource descriptor.
func (s *Sanitizer) GVR() client.GVR { return s.gvr }

// App returns the current app handle.
func (s *Sanitizer) App() *App {
	return s.app
}

// UpdateTitle updates the view title.
func (s *Sanitizer) UpdateTitle() {
	t := s.styleTitle()
	s.app.QueueUpdateDraw(func() {
		s.SetTitle(t)
	})
}

func (s *Sanitizer) styleTitle() string {
	base := strings.Title(s.gvr.R())
	ns := s.model.GetNamespace()
	if client.IsAllNamespaces(ns) {
		ns = client.NamespaceAll
	}

	buff := s.CmdBuff().String()
	var title string
	if ns == client.ClusterScope {
		title = ui.SkinTitle(fmt.Sprintf(ui.TitleFmt, base, s.Count), s.app.Styles.Frame())
	} else {
		title = ui.SkinTitle(fmt.Sprintf(ui.NSTitleFmt, base, ns, s.Count), s.app.Styles.Frame())
	}
	if buff == "" {
		return title
	}

	if ui.IsLabelSelector(buff) {
		buff = ui.TrimLabelSelector(buff)
	}

	return title + ui.SkinTitle(fmt.Sprintf(ui.SearchFmt, buff), s.app.Styles.Frame())
}
