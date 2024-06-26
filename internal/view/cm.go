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

// ConfigMap represents a configmap viewer.
type ConfigMap struct {
	ResourceViewer
}

// NewConfigMap returns a new viewer.
func NewConfigMap(gvr client.GVR) ResourceViewer {
	s := ConfigMap{
		ResourceViewer: NewOwnerExtender(
			NewBrowser(gvr),
		),
	}
	s.AddBindKeysFn(s.bindKeys)

	return &s
}

func (s *ConfigMap) bindKeys(aa *ui.KeyActions) {
	aa.Add(ui.KeyU, ui.NewKeyAction("UsedBy", s.refCmd, true))
}

func (s *ConfigMap) refCmd(evt *tcell.EventKey) *tcell.EventKey {
	return scanRefs(evt, s.App(), s.GetTable(), dao.CmGVR)
}

func scanRefs(evt *tcell.EventKey, a *App, t *Table, gvr client.GVR) *tcell.EventKey {
	path := t.GetSelectedItem()
	if path == "" {
		return evt
	}

	ctx := context.Background()
	refs, err := dao.ScanForRefs(refContext(gvr, path, true)(ctx), a.factory)
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

func refContext(gvr client.GVR, path string, wait bool) ContextFunc {
	return func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, internal.KeyPath, path)
		ctx = context.WithValue(ctx, internal.KeyGVR, gvr)
		return context.WithValue(ctx, internal.KeyWait, wait)
	}
}
