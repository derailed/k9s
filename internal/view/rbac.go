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

// Rbac presents an RBAC policy viewer.
type Rbac struct {
	ResourceViewer
}

// NewRbac returns a new viewer.
func NewRbac(gvr client.GVR) ResourceViewer {
	r := Rbac{
		ResourceViewer: NewBrowser(gvr),
	}
	r.AddBindKeysFn(r.bindKeys)
	r.GetTable().SetSortCol("API-GROUP", true)
	r.GetTable().SetEnterFn(blankEnterFn)

	return &r
}

func (r *Rbac) bindKeys(aa *ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyShiftA, ui.NewKeyAction("Sort API-Group", r.GetTable().SortColCmd("API-GROUP", true), false))
}

func showRules(app *App, _ ui.Tabular, gvr client.GVR, path string) {
	v := NewRbac(client.NewGVR("rbac"))
	v.SetContextFn(rbacCtx(gvr, path))

	if err := app.inject(v, false); err != nil {
		app.Flash().Err(err)
	}
}

func rbacCtx(gvr client.GVR, path string) ContextFunc {
	return func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, internal.KeyPath, path)
		return context.WithValue(ctx, internal.KeyGVR, gvr)
	}
}

func blankEnterFn(_ *App, _ ui.Tabular, _ client.GVR, _ string) {}
