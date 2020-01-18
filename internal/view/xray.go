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
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/xray"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	"github.com/sahilm/fuzzy"
)

const xrayTitle = "Xray"

// Xray represents an xray tree view.
type Xray struct {
	*tview.TreeView

	actions      ui.KeyActions
	app          *App
	gvr          client.GVR
	selectedNode string
	model        *model.Tree
	cancelFn     context.CancelFunc
	cmdBuff      *ui.CmdBuff
	expandNodes  bool
}

var _ ResourceViewer = (*Xray)(nil)

// NewXray returns a new view.
func NewXray(gvr client.GVR) ResourceViewer {
	a := Xray{
		TreeView:    tview.NewTreeView(),
		model:       model.NewTree(gvr.String()),
		expandNodes: true,
		actions:     make(ui.KeyActions),
		cmdBuff:     ui.NewCmdBuff('/', ui.FilterBuff),
	}

	return &a
}

// Init initializes the view
func (x *Xray) Init(ctx context.Context) error {
	var err error
	if x.app, err = extractApp(ctx); err != nil {
		return err
	}

	x.bindKeys()
	x.SetBorder(true)
	x.SetBorderAttributes(tcell.AttrBold)
	x.SetBorderPadding(0, 0, 1, 1)
	x.SetBackgroundColor(config.AsColor(x.app.Styles.GetTable().BgColor))
	x.SetBorderColor(config.AsColor(x.app.Styles.GetTable().FgColor))
	x.SetBorderFocusColor(config.AsColor(x.app.Styles.Frame().Border.FocusColor))
	x.SetTitle(" Xray ")
	x.SetGraphics(true)
	x.SetGraphicsColor(tcell.ColorDimGray)
	x.SetInputCapture(x.keyboard)

	x.model.SetRefreshRate(time.Duration(x.app.Config.K9s.GetRefreshRate()) * time.Second)
	x.model.SetNamespace(client.AllNamespaces)
	x.model.AddListener(x)

	x.SetChangedFunc(func(n *tview.TreeNode) {
		ref, ok := n.GetReference().(xray.NodeSpec)
		if !ok {
			log.Error().Msgf("No ref found on node %s", n.GetText())
			return
		}
		x.selectedNode = ref.Path
	})

	return nil
}

// SetInstance sets specific resource instance.
func (x *Xray) SetInstance(string) {}

// Actions returns active menu bindings.
func (x *Xray) Actions() ui.KeyActions {
	return x.actions
}

// Hints returns the view hints.
func (x *Xray) Hints() model.MenuHints {
	return x.actions.Hints()
}

func (x *Xray) bindKeys() {
	x.Actions().Add(ui.KeyActions{
		ui.KeySpace:         ui.NewKeyAction("Expand/Collapse", x.noopCmd, true),
		ui.KeyE:             ui.NewKeyAction("Expand/Collapse All", x.toggleCollapseCmd, true),
		ui.KeyV:             ui.NewKeyAction("Goto", x.gotoCmd, true),
		tcell.KeyEnter:      ui.NewKeyAction("Goto", x.gotoCmd, true),
		ui.KeySlash:         ui.NewSharedKeyAction("Filter Mode", x.activateCmd, false),
		tcell.KeyBackspace2: ui.NewSharedKeyAction("Erase", x.eraseCmd, false),
		tcell.KeyBackspace:  ui.NewSharedKeyAction("Erase", x.eraseCmd, false),
		tcell.KeyDelete:     ui.NewSharedKeyAction("Erase", x.eraseCmd, false),
		tcell.KeyCtrlU:      ui.NewSharedKeyAction("Clear Filter", x.clearCmd, false),
		tcell.KeyEscape:     ui.NewSharedKeyAction("Filter Reset", x.resetCmd, false),
	})
}

func (x *Xray) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		if x.cmdBuff.IsActive() {
			x.cmdBuff.Add(evt.Rune())
			x.ClearSelection()
			x.update(x.filter(x.model.Peek()))
			x.UpdateTitle()
			return nil
		}

		key = mapKey(evt)
	}

	if a, ok := x.actions[key]; ok {
		return a.Action(evt)
	}

	return evt
}

func (x *Xray) noopCmd(evt *tcell.EventKey) *tcell.EventKey {
	return evt
}

func (x *Xray) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if x.app.InCmdMode() {
		return evt
	}
	x.app.Flash().Info("Filter mode activated.")
	x.cmdBuff.SetActive(true)

	return nil
}

func (x *Xray) clearCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !x.cmdBuff.IsActive() {
		return evt
	}
	x.cmdBuff.Clear()
	x.model.ClearFilter()
	x.Start()

	return nil
}

func (x *Xray) eraseCmd(evt *tcell.EventKey) *tcell.EventKey {
	if x.cmdBuff.IsActive() {
		x.cmdBuff.Delete()
	}
	x.UpdateTitle()

	return nil
}

func (x *Xray) filterCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !x.cmdBuff.IsActive() {
		return evt
	}
	x.cmdBuff.SetActive(false)

	cmd := x.cmdBuff.String()
	x.model.SetFilter(cmd)
	x.Start()

	return nil
}

func (x *Xray) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !x.cmdBuff.InCmdMode() {
		x.cmdBuff.Reset()
		return x.app.PrevCmd(evt)
	}

	x.app.Flash().Info("Clearing filter...")
	x.cmdBuff.Reset()
	x.model.ClearFilter()
	x.Start()

	return nil
}

func (x *Xray) gotoCmd(evt *tcell.EventKey) *tcell.EventKey {
	if x.cmdBuff.IsActive() {
		if ui.IsLabelSelector(x.cmdBuff.String()) {
			x.Start()
			return nil
		}
	}
	n := x.GetCurrentNode()
	if n == nil {
		return nil
	}

	ref, ok := n.GetReference().(xray.NodeSpec)
	if !ok {
		log.Error().Msgf("Expecting a NodeSpec!")
		return nil
	}
	if len(strings.Split(ref.Path, "/")) == 1 {
		return nil
	}

	if err := x.app.viewResource(client.NewGVR(ref.GVR).ToR(), ref.Path, false); err != nil {
		x.app.Flash().Err(err)
	}
	return nil
}

func (x *Xray) toggleCollapseCmd(evt *tcell.EventKey) *tcell.EventKey {
	x.expandNodes = !x.expandNodes
	x.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		if parent != nil {
			node.SetExpanded(x.expandNodes)
		}
		return true
	})
	return nil
}

// ClearSelection clears the currently selected node.
func (x *Xray) ClearSelection() {
	x.selectedNode = ""
	x.SetCurrentNode(nil)
}

func (x *Xray) filter(root *xray.TreeNode) *xray.TreeNode {
	q := x.cmdBuff.String()
	if x.cmdBuff.Empty() || ui.IsLabelSelector(q) {
		return root
	}

	x.UpdateTitle()
	if ui.IsFuzzySelector(q) {
		return root.Filter(q, fuzzyFilter)
	}

	return root.Filter(q, rxFilter)
}

// TreeNodeSelected callback for node selection.
func (x *Xray) TreeNodeSelected() {
	x.app.QueueUpdateDraw(func() {
		n := x.GetCurrentNode()
		if n != nil {
			n.SetColor(config.AsColor(x.app.Styles.GetTable().CursorColor))
		}
	})
}

// XrayLoadFailed notifies the load failed.
func (x *Xray) TreeLoadFailed(err error) {
	x.app.Flash().Err(err)
}

func (x *Xray) update(node *xray.TreeNode) {
	root := makeTreeNode(node, x.expandNodes, x.app.Styles)
	if node == nil {
		x.app.QueueUpdateDraw(func() {
			x.SetRoot(root)
		})
		return
	}

	for _, c := range node.Children {
		x.hydrate(root, c)
	}
	if x.selectedNode == "" {
		x.selectedNode = node.ID
	}

	x.app.QueueUpdateDraw(func() {
		x.SetRoot(root)
		root.Walk(func(node, parent *tview.TreeNode) bool {
			ref := node.GetReference().(xray.NodeSpec)
			// BOZO!! Figure this out expand/collapse but the root
			if parent != nil {
				node.SetExpanded(x.expandNodes)
			} else {
				node.SetExpanded(true)
			}

			ref, ok := node.GetReference().(xray.NodeSpec)
			if !ok {
				log.Error().Msgf("No ref found on node %s", node.GetText())
				return false
			}
			if ref.Path == x.selectedNode {
				node.SetExpanded(true).SetSelectable(true)
				x.SetCurrentNode(node)
			}
			return true
		})
	})
}

// XrayDataChanged notifies the model data changed.
func (x *Xray) TreeChanged(node *xray.TreeNode) {
	log.Debug().Msgf("Tree Changed %d", len(node.Children))
	x.update(x.filter(node))
	x.UpdateTitle()
}

func (x *Xray) hydrate(parent *tview.TreeNode, n *xray.TreeNode) {
	node := makeTreeNode(n, x.expandNodes, x.app.Styles)
	for _, c := range n.Children {
		x.hydrate(node, c)
	}
	parent.AddChild(node)
}

// SetEnvFn sets the custom environment function.
func (x *Xray) SetEnvFn(EnvFunc) {}

// Refresh refresh the view
func (x *Xray) Refresh() {
}

// BufferChanged indicates the buffer was changed.
func (x *Xray) BufferChanged(s string) {}

// BufferActive indicates the buff activity changed.
func (x *Xray) BufferActive(state bool, k ui.BufferKind) {
	x.app.BufferActive(state, k)
}

func (x *Xray) defaultContext() context.Context {
	ctx := context.WithValue(context.Background(), internal.KeyFactory, x.app.factory)
	ctx = context.WithValue(ctx, internal.KeyFields, "")
	if x.cmdBuff.Empty() {
		ctx = context.WithValue(ctx, internal.KeyLabels, "")
	} else {
		ctx = context.WithValue(ctx, internal.KeyLabels, ui.TrimLabelSelector(x.cmdBuff.String()))
	}

	return ctx
}

// Start initializes resource watch loop.
func (x *Xray) Start() {
	x.Stop()

	log.Debug().Msgf("XRAY STARTING! -- %q", x.selectedNode)
	x.cmdBuff.AddListener(x.app.Cmd())
	x.cmdBuff.AddListener(x)
	x.app.SetFocus(x)

	ctx := x.defaultContext()
	ctx, x.cancelFn = context.WithCancel(ctx)
	x.model.Watch(ctx)
	x.UpdateTitle()
}

// Stop terminates watch loop.
func (x *Xray) Stop() {
	log.Debug().Msgf("XRAY STOPPING!")
	if x.cancelFn == nil {
		return
	}
	x.cancelFn()
	x.cancelFn = nil

	x.cmdBuff.RemoveListener(x.app.Cmd())
	x.cmdBuff.RemoveListener(x)
}

// SetBindKeysFn sets up extra key bindings.
func (x *Xray) SetBindKeysFn(BindKeysFunc) {}

// SetContextFn sets custom context.
func (x *Xray) SetContextFn(ContextFunc) {}

// Name returns the component name.
func (x *Xray) Name() string { return "XRay" }

// GetTable returns the underlying table.
func (x *Xray) GetTable() *Table { return nil }

// GVR returns a resource descriptor.
func (x *Xray) GVR() string { return x.gvr.String() }

// App returns the current app handle.
func (x *Xray) App() *App {
	return x.app
}

// UpdateTitle updates the view title.
func (x *Xray) UpdateTitle() {
	x.SetTitle(x.styleTitle())
}

func (x *Xray) styleTitle() string {
	rc := x.GetRowCount()
	if rc > 0 {
		rc--
	}

	base := strings.Title(xrayTitle)
	ns := x.model.GetNamespace()
	if client.IsAllNamespaces(ns) {
		ns = client.NamespaceAll
	}

	buff := x.cmdBuff.String()
	var title string
	if ns == client.ClusterScope {
		title = ui.SkinTitle(fmt.Sprintf(ui.TitleFmt, base, rc), x.app.Styles.Frame())
	} else {
		title = ui.SkinTitle(fmt.Sprintf(ui.NSTitleFmt, base, ns, rc), x.app.Styles.Frame())
	}
	if buff == "" {
		return title
	}

	if ui.IsLabelSelector(buff) {
		buff = ui.TrimLabelSelector(buff)
	}

	return title + ui.SkinTitle(fmt.Sprintf(ui.SearchFmt, buff), x.app.Styles.Frame())
}

// ----------------------------------------------------------------------------
// Helpers...

func mapKey(evt *tcell.EventKey) tcell.Key {
	key := tcell.Key(evt.Rune())
	if evt.Modifiers() == tcell.ModAlt {
		key = tcell.Key(int16(evt.Rune()) * int16(evt.Modifiers()))
	}
	return key
}

func fuzzyFilter(q, path string) bool {
	q = strings.TrimSpace(q[2:])
	mm := fuzzy.Find(q, []string{path})
	log.Debug().Msgf("%#v", mm)
	if len(mm) > 0 {
		return true
	}

	return false
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

func makeTreeNode(node *xray.TreeNode, expanded bool, styles *config.Styles) *tview.TreeNode {
	n := tview.NewTreeNode("No data...")
	if node != nil {
		n.SetText(node.Title())
		n.SetReference(xray.NodeSpec{GVR: node.GVR, Path: node.ID})
	}
	n.SetSelectable(true)
	n.SetExpanded(expanded)
	n.SetColor(config.AsColor(styles.GetTable().CursorColor))
	n.SetSelectedFunc(func() {
		n.SetExpanded(!n.IsExpanded())
	})
	return n
}
