// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"testing"

	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTreeToggleCollapse(t *testing.T) {
	uu := map[string]struct {
		root *tview.TreeNode
	}{
		"no-root": {},
		"root": {
			root: tview.NewTreeNode("root").AddChild(tview.NewTreeNode("child")),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			v := ui.NewTree()
			v.BindKeys()
			if u.root != nil {
				v.SetRoot(u.root)
			}
			a, ok := v.Actions().Get(ui.KeyX)
			require.True(t, ok)

			assert.NotPanics(t, func() {
				a.Action(tcell.NewEventKey(tcell.KeyRune, 'x', tcell.ModNone))
			})
			assert.False(t, v.ExpandNodes())
			if u.root != nil {
				assert.False(t, u.root.GetChildren()[0].IsExpanded())
			}
		})
	}
}
