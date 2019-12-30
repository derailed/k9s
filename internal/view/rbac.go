package view

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
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
	r.GetTable().SetColorerFn(render.Rbac{}.ColorerFunc())
	r.SetBindKeysFn(r.bindKeys)
	r.GetTable().SetSortCol(1, len(render.Rbac{}.Header(render.ClusterScope)), true)
	r.GetTable().SetEnterFn(blankEnterFn)

	return &r
}

func (r *Rbac) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyActions{
		ui.KeyShiftO: ui.NewKeyAction("Sort APIGroup", r.GetTable().SortColCmd(1, true), false),
	})
}

func showRules(app *App, _, gvr, path string) {
	v := NewRbac(client.GVR("rbac"))
	v.SetContextFn(rbacCtxt(gvr, path))

	if err := app.inject(v); err != nil {
		app.Flash().Err(err)
	}
}

func rbacCtxt(gvr, path string) ContextFunc {
	return func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, internal.KeyPath, path)
		return context.WithValue(ctx, internal.KeyGVR, gvr)
	}
}

func blankEnterFn(_ *App, _, _, _ string) {}
