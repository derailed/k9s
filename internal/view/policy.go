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

const (
	group = "Group"
	user  = "User"
	sa    = "ServiceAccount"
)

// Policy presents a RBAC rules viewer based on what a given user/group or sa can do.
type Policy struct {
	ResourceViewer

	subjectKind, subjectName string
}

// NewPolicy returns a new viewer.
func NewPolicy(_ *App, subject, name string) *Policy {
	p := Policy{
		ResourceViewer: NewBrowser(client.PolGVR),
		subjectKind:    subject,
		subjectName:    name,
	}
	p.AddBindKeysFn(p.bindKeys)
	p.GetTable().SetSortCol("API-GROUP", false)
	p.SetContextFn(p.subjectCtx)
	p.GetTable().SetEnterFn(blankEnterFn)

	return &p
}

func (p *Policy) subjectCtx(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, internal.KeySubjectKind, mapSubject(p.subjectKind))
	ctx = context.WithValue(ctx, internal.KeyPath, mapSubject(p.subjectKind)+":"+p.subjectName)
	return context.WithValue(ctx, internal.KeySubjectName, p.subjectName)
}

func (*Policy) bindKeys(aa *ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlSpace, ui.KeySpace)
}

func mapSubject(subject string) string {
	switch subject {
	case "g":
		return group
	case "s":
		return sa
	case "u":
		return user
	default:
		return subject
	}
}
