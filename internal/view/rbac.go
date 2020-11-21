package view

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell/v2"
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
	r.AddBindKeysFn(r.bindKeys)
	r.GetTable().SetSortCol("APIGROUP", true)
	r.GetTable().SetEnterFn(blankEnterFn)

	return &r
}

func (r *Rbac) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyActions{
		ui.KeyShiftO: ui.NewKeyAction("Sort APIGroup", r.GetTable().SortColCmd("APIGROUP", true), false),
	})
}

func showRules(app *App, _ ui.Tabular, gvr, path string) {
	v := NewRbac(client.NewGVR("rbac"))
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

func blankEnterFn(_ *App, _ ui.Tabular, _, _ string) {}
