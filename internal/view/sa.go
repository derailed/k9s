// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

// ServiceAccount represents a serviceaccount viewer.
type ServiceAccount struct {
	ResourceViewer
}

// NewServiceAccount returns a new viewer.
func NewServiceAccount(gvr client.GVR) ResourceViewer {
	s := ServiceAccount{
		ResourceViewer: NewBrowser(gvr),
	}
	s.AddBindKeysFn(s.bindKeys)
	s.SetContextFn(s.subjectCtx)

	return &s
}

func (s *ServiceAccount) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyU:        ui.NewKeyAction("UsedBy", s.refCmd, true),
		tcell.KeyEnter: ui.NewKeyAction("Rules", s.policyCmd, true),
	})
}

func (s *ServiceAccount) subjectCtx(ctx context.Context) context.Context {
	return context.WithValue(ctx, internal.KeySubjectKind, sa)
}

func (s *ServiceAccount) refCmd(evt *tcell.EventKey) *tcell.EventKey {
	return scanSARefs(evt, s.App(), s.GetTable(), dao.SaGVR)
}

func (s *ServiceAccount) policyCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := s.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}
	if err := s.App().inject(NewPolicy(s.App(), sa, path), false); err != nil {
		s.App().Flash().Err(err)
	}

	return nil
}

func scanSARefs(evt *tcell.EventKey, a *App, t *Table, gvr client.GVR) *tcell.EventKey {
	path := t.GetSelectedItem()
	if path == "" {
		return evt
	}

	ctx := context.Background()
	refs, err := dao.ScanForSARefs(refContext(gvr, path, true)(ctx), a.factory)
	if err != nil {
		a.Flash().Err(err)
		return nil
	}
	if len(refs) == 0 {
		a.Flash().Warnf("No references found at this time for %s::%s. Check again later!", gvr, path)
		return nil
	}
	a.Flash().Infof("Viewing references for %s::%s", gvr, path)
	view := NewReference(client.NewGVR("references"))
	view.SetContextFn(refContext(gvr, path, false))
	if err := a.inject(view, false); err != nil {
		a.Flash().Err(err)
	}

	return nil
}
