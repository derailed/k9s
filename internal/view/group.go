// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

// Group presents a RBAC group viewer.
type Group struct {
	ResourceViewer
}

// NewGroup returns a new subject viewer.
func NewGroup(gvr client.GVR) ResourceViewer {
	g := Group{ResourceViewer: NewBrowser(gvr)}
	g.AddBindKeysFn(g.bindKeys)
	g.SetContextFn(g.subjectCtx)

	return &g
}

func (g *Group) bindKeys(aa *ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, ui.KeyShiftP, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Bulk(ui.KeyMap{
		tcell.KeyEnter: ui.NewKeyAction("Rules", g.policyCmd, true),
		ui.KeyShiftK:   ui.NewKeyAction("Sort Kind", g.GetTable().SortColCmd("KIND", true), false),
	})
}

func (g *Group) subjectCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeySubjectKind, "Group")
}

func (g *Group) policyCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := g.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}
	if err := g.App().inject(NewPolicy(g.App(), "Group", path), false); err != nil {
		g.App().Flash().Err(err)
	}

	return nil
}
