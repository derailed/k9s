package view

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

// User presents a user viewer.
type User struct {
	ResourceViewer
}

// NewUser returns a new subject viewer.
func NewUser(gvr client.GVR) ResourceViewer {
	s := User{ResourceViewer: NewBrowser(gvr)}
	s.GetTable().SetColorerFn(render.Subject{}.ColorerFunc())
	s.SetBindKeysFn(s.bindKeys)
	s.SetContextFn(s.subjectCtx)
	return &s
}

func (s *User) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, ui.KeyShiftP, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyActions{
		tcell.KeyEnter: ui.NewKeyAction("Rules", s.policyCmd, true),
		ui.KeyShiftK:   ui.NewKeyAction("Sort Kind", s.GetTable().SortColCmd(1, true), false),
	})
}

func (s *User) subjectCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeySubjectKind, "User")
}

func (s *User) policyCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !s.GetTable().RowSelected() {
		return evt
	}
	if err := s.App().inject(NewPolicy(s.App(), "User", s.GetTable().GetSelectedItem())); err != nil {
		s.App().Flash().Err(err)
	}

	return nil
}
