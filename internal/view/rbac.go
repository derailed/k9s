package view

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	ClusterRole roleKind = iota
	Role
)

var (
	k8sVerbs = []string{
		"get",
		"list",
		"watch",
		"create",
		"patch",
		"update",
		"delete",
		"deletecollection",
	}

	httpTok8sVerbs = map[string]string{
		"post": "create",
		"put":  "update",
	}
)

type roleKind = int8

// Rbac presents an RBAC policy viewer.
type Rbac struct {
	ResourceViewer
}

// NewRbac returns a new viewer.
func NewRbac(gvr client.GVR) ResourceViewer {
	log.Debug().Msgf(">>>>> NEWRBAC %v!!!!!", gvr)
	r := Rbac{
		ResourceViewer: NewBrowser(gvr),
	}
	r.GetTable().SetColorerFn(render.Rbac{}.ColorerFunc())
	r.SetBindKeysFn(r.bindKeys)
	r.GetTable().SetSortCol(1, len(render.Rbac{}.Header(render.ClusterScope)), true)

	return &r
}

func (r *Rbac) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyActions{
		ui.KeyShiftO: ui.NewKeyAction("Sort APIGroup", r.GetTable().SortColCmd(1, true), false),
	})
}

// BOZO!!
// func showClusterRoleBinding(app *App, ns, gvr, path string) {
// 	o, err := app.factory.Get("rbac.authorization.k8s.io/v1/clusterrolebindings", path, labels.Everything())
// 	if err != nil {
// 		app.Flash().Err(err)
// 		return
// 	}

// 	var crb rbacv1.ClusterRoleBinding
// 	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &crb)
// 	if err != nil {
// 		app.Flash().Errf("Unable to retrieve clusterrolebindings for %s", path)
// 		return
// 	}

// 	// BOZO!! Must make sure cluster roles are in cache prior to loading rbac view.
// 	app.factory.ForResource("-", "rbac.authorization.k8s.io/v1/clusterroles")
// 	app.factory.WaitForCacheSync()

// 	// BOZO!!
// 	// app.inject(NewRbac(crb.RoleRef.Name, ClusterRole, selection))
// }

func showRBAC(app *App, _, gvr, path string) {
	log.Debug().Msgf("Showing RBAC %q--%q", gvr, path)
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
