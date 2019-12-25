package view

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

type (
	TableInfo interface {
		Header() render.HeaderRow
		GetCache() render.RowEvents
		SetCache(render.RowEvents)
	}

	// Subject presents a user/group viewer.
	Subject struct {
		ResourceViewer

		subjectKind string
	}
)

// NewSubject returns a new subject viewer.
func NewSubject(gvr client.GVR) ResourceViewer {
	s := Subject{ResourceViewer: NewBrowser(gvr)}
	s.GetTable().SetColorerFn(render.Subject{}.ColorerFunc())
	// BOZO!!
	// s.GetTable().SetSortCol(1, len(s.Header()), true)
	s.SetBindKeysFn(s.bindKeys)
	s.SetContextFn(s.subjectCtx)
	return &s
}

// Name returns the component name
func (s *Subject) Name() string {
	return "subjects"
}

func (s *Subject) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, ui.KeyShiftP, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyActions{
		tcell.KeyEnter: ui.NewKeyAction("Policies", s.policyCmd, true),
		ui.KeyShiftK:   ui.NewKeyAction("Sort Kind", s.GetTable().SortColCmd(1, true), false),
	})
}

func (s *Subject) subjectCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeySubjectKind, mapSubject(s.subjectKind))
}

// SetSubject sets the subject name.
func (s *Subject) SetSubject(n string) {
	s.subjectKind = mapSubject(n)
}

func (s *Subject) policyCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !s.GetTable().RowSelected() {
		return evt
	}

	// BOZO!!
	// _, n := client.Namespaced(s.GetSelectedItem())
	// subject, err := mapFuSubject(s.subjectKind)
	// if err != nil {
	// 	s.App().Flash().Err(err)
	// 	return nil
	// }
	// BOZO!!
	// s.App().inject(NewPolicy(s.app, subject, n))

	return nil
}
