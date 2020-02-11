package ui

import (
	"context"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

// KeyListenerFunc listens to key presses.
type KeyListenerFunc func()

// Tree represents a tree view.
type Tree struct {
	*tview.TreeView

	actions      KeyActions
	selectedItem string
	cmdBuff      *CmdBuff
	expandNodes  bool
	Count        int
	keyListener  KeyListenerFunc
}

// NewTree returns a new view.
func NewTree() *Tree {
	return &Tree{
		TreeView:    tview.NewTreeView(),
		expandNodes: true,
		actions:     make(KeyActions),
		cmdBuff:     NewCmdBuff('/', FilterBuff),
	}
}

// Init initializes the view
func (t *Tree) Init(ctx context.Context) error {
	t.BindKeys()
	t.SetBorder(true)
	t.SetBorderAttributes(tcell.AttrBold)
	t.SetBorderPadding(0, 0, 1, 1)
	t.SetGraphics(true)
	t.SetGraphicsColor(tcell.ColorFloralWhite)
	t.SetInputCapture(t.keyboard)

	return nil
}

// SetSelectedItem sets the currently selected node.
func (t *Tree) SetSelectedItem(s string) {
	t.selectedItem = s
}

// GetSelectedItem returns the currently selected item or blank if none.
func (t *Tree) GetSelectedItem() string {
	return t.selectedItem
}

// ExpandNodes returns true if nodes are expanded or false otherwise.
func (t *Tree) ExpandNodes() bool {
	return t.expandNodes
}

// CmdBuff returns the filter command.
func (t *Tree) CmdBuff() *CmdBuff {
	return t.cmdBuff
}

// SetKeyListenerFn sets a key entered listener.
func (t *Tree) SetKeyListenerFn(f KeyListenerFunc) {
	t.keyListener = f
}

// Actions returns active menu bindings.
func (t *Tree) Actions() KeyActions {
	return t.actions
}

// Hints returns the view hints.
func (t *Tree) Hints() model.MenuHints {
	return t.actions.Hints()
}

// ExtraHints returns additional hints.
func (t *Tree) ExtraHints() map[string]string {
	return nil
}

func (t *Tree) BindKeys() {
	t.Actions().Add(KeyActions{
		KeySpace: NewKeyAction("Expand/Collapse", t.noopCmd, true),
		KeyX:     NewKeyAction("Expand/Collapse All", t.toggleCollapseCmd, true),
	})
}

func (t *Tree) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		if t.cmdBuff.IsActive() {
			t.cmdBuff.Add(evt.Rune())
			t.ClearSelection()
			if t.keyListener != nil {
				t.keyListener()
			}
			return nil
		}
		key = mapKey(evt)
	}

	if a, ok := t.actions[key]; ok {
		return a.Action(evt)
	}

	return evt
}

func (t *Tree) noopCmd(evt *tcell.EventKey) *tcell.EventKey {
	return evt
}

func (t *Tree) toggleCollapseCmd(evt *tcell.EventKey) *tcell.EventKey {
	t.expandNodes = !t.expandNodes
	t.GetRoot().Walk(func(node, parent *tview.TreeNode) bool {
		if parent != nil {
			node.SetExpanded(t.expandNodes)
		}
		return true
	})
	return nil
}

// ClearSelection clears the currently selected node.
func (t *Tree) ClearSelection() {
	t.selectedItem = ""
	t.SetCurrentNode(nil)
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
