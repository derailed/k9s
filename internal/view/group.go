package view

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

// Group presents a RBAC group viewer.
type Group struct {
	ResourceViewer
}

// NewGroup returns a new subject viewer.
func NewGroup(gvr client.GVR) ResourceViewer {
	s := Group{ResourceViewer: NewBrowser(gvr)}
	s.GetTable().SetColorerFn(render.Subject{}.ColorerFunc())
	s.SetBindKeysFn(s.bindKeys)
	s.SetContextFn(s.subjectCtx)
	return &s
}

func (s *Group) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, ui.KeyShiftP, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyActions{
		tcell.KeyEnter: ui.NewKeyAction("Rules", s.policyCmd, true),
		ui.KeyShiftK:   ui.NewKeyAction("Sort Kind", s.GetTable().SortColCmd(1, true), false),
	})
}

func (s *Group) subjectCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeySubjectKind, "Group")
}

func (s *Group) policyCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !s.GetTable().RowSelected() {
		return evt
	}
	if err := s.App().inject(NewPolicy(s.App(), "Group", s.GetTable().GetSelectedItem())); err != nil {
		s.App().Flash().Err(err)
	}

	return nil
}
