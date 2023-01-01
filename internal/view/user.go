package view

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell/v2"
)

// User presents a user viewer.
type User struct {
	ResourceViewer
}

// NewUser returns a new subject viewer.
func NewUser(gvr client.GVR) ResourceViewer {
	u := User{ResourceViewer: NewBrowser(gvr)}
	u.AddBindKeysFn(u.bindKeys)
	u.SetContextFn(u.subjectCtx)

	return &u
}

func (u *User) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, ui.KeyShiftP, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyActions{
		tcell.KeyEnter: ui.NewKeyAction("Rules", u.policyCmd, true),
		ui.KeyShiftK:   ui.NewKeyAction("Sort Kind", u.GetTable().SortColCmd("KIND", true), false),
	})
}

func (u *User) subjectCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeySubjectKind, "User")
}

func (u *User) policyCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := u.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}
	if err := u.App().inject(NewPolicy(u.App(), "User", path)); err != nil {
		u.App().Flash().Err(err)
	}

	return nil
}
