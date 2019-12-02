package view

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	ClusterRole roleKind = iota
	Role

	clusterWide  = "*"
	rbacTitle    = "Rbac"
	rbacTitleFmt = " [fg:bg:b]%s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-][[count:bg:b]%d[fg:bg:-]][fg:bg:-] "
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
	*Table

	roleType roleKind
	roleName string
	path     string
	cache    render.RowEvents
}

// NewRbac returns a new viewer.
func NewRbac(name string, kind roleKind, path string) *Rbac {
	return &Rbac{
		Table:    NewTable(rbacTitle),
		roleName: name,
		roleType: kind,
		path:     path,
	}
}

// Init initializes the view.
func (r *Rbac) Init(ctx context.Context) error {
	if err := r.Table.Init(ctx); err != nil {
		return err
	}
	r.SetColorerFn(render.Rbac{}.ColorerFunc())
	r.bindKeys()
	r.SetSortCol(1, len(r.Header()), true)
	r.refresh()

	return nil
}

func (r *Rbac) UpdateTitle() {
	r.SetTitle(ui.SkinTitle(fmt.Sprintf(rbacTitleFmt, rbacTitle, r.path, r.GetRowCount()-1), r.app.Styles.Frame()))
}

// Start watches for viewer updates
func (r *Rbac) Start() {
	if r.app.Conn() == nil {
		return
	}

	r.Stop()

	var ctx context.Context
	ctx, r.cancelFn = context.WithCancel(context.Background())

	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(r.app.Config.K9s.GetRefreshRate()) * time.Second):
				r.app.QueueUpdateDraw(func() {
					r.refresh()
				})
			}
		}
	}(ctx)
}

// Name returns the component name.
func (r *Rbac) Name() string {
	return rbacTitle
}

func (r *Rbac) bindKeys() {
	r.Actions().Delete(ui.KeyShiftA, tcell.KeyCtrlSpace, ui.KeySpace)
	r.Actions().Add(ui.KeyActions{
		tcell.KeyEscape: ui.NewKeyAction("Reset", r.resetCmd, false),
		ui.KeySlash:     ui.NewKeyAction("Filter", r.activateCmd, false),
		ui.KeyShiftO:    ui.NewKeyAction("Sort APIGroup", r.SortColCmd(1, true), false),
	})
}

func (r *Rbac) refresh() {
	if r.app.Conn() == nil {
		return
	}
	data, err := r.reconcile(r.roleName, r.roleType)
	if err != nil {
		log.Error().Err(err).Msgf("Refresh for %s:%d", r.roleName, r.roleType)
		r.app.Flash().Err(err)
	}
	r.Update(data)
	r.UpdateTitle()
}

func (r *Rbac) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.SearchBuff().Empty() {
		r.SearchBuff().Reset()
		return nil
	}

	return r.backCmd(evt)
}

func (r *Rbac) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if r.cancelFn != nil {
		r.cancelFn()
	}

	if r.SearchBuff().IsActive() {
		r.SearchBuff().Reset()
		return nil
	}

	return r.app.PrevCmd(evt)
}

func (r *Rbac) reconcile(name string, kind roleKind) (render.TableData, error) {
	var table render.TableData

	rows, err := r.fetchRoles(name, kind)
	if err != nil {
		return table, err
	}

	return buildTable(r, rows), nil
}

func (r *Rbac) Header() render.HeaderRow {
	return render.Rbac{}.Header(render.AllNamespaces)
}

func (r *Rbac) GetCache() render.RowEvents {
	return r.cache
}

func (r *Rbac) SetCache(evts render.RowEvents) {
	r.cache = evts
}

func (r *Rbac) fetchRoles(name string, kind roleKind) (render.Rows, error) {
	switch kind {
	case ClusterRole:
		return r.loadClusterRoles(name)
	case Role:
		return r.loadRoles(name)
	default:
		return nil, fmt.Errorf("Expecting clusterrole/role but found %d", kind)
	}
}

func (r *Rbac) loadClusterRoles(name string) (render.Rows, error) {
	o, err := r.app.factory.Get("-", "rbac.authorization.k8s.io/v1/clusterroles", name, labels.Everything())
	if err != nil {
		return nil, err
	}

	var cr rbacv1.ClusterRole
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cr)
	if err != nil {
		return nil, err
	}

	return r.parseRules(cr.Rules), nil
}

func (r *Rbac) loadRoles(path string) (render.Rows, error) {
	ns, n := namespaced(path)
	o, err := r.app.factory.Get(ns, "rbac.authorization.k8s.io/v1/roles", n, labels.Everything())
	if err != nil {
		return nil, err
	}

	var ro rbacv1.Role
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ro)
	if err != nil {
		return nil, err
	}

	return r.parseRules(ro.Rules), nil
}

func (r *Rbac) parseRules(rules []rbacv1.PolicyRule) render.Rows {
	m := make(render.Rows, 0, len(rules))
	for _, rule := range rules {
		for _, grp := range rule.APIGroups {
			for _, res := range rule.Resources {
				k := res
				if grp != "" {
					k = res + "." + grp
				}
				for _, na := range rule.ResourceNames {
					m = m.Upsert(r.prepRow(fqn(k, na), grp, rule.Verbs))
				}
				m = m.Upsert(r.prepRow(k, grp, rule.Verbs))
			}
		}
		for _, nres := range rule.NonResourceURLs {
			if nres[0] != '/' {
				nres = "/" + nres
			}
			m = m.Upsert(r.prepRow(nres, "", rule.Verbs))
		}
	}

	return m
}

func (r *Rbac) prepRow(res, grp string, verbs []string) render.Row {
	if grp != "" {
		grp = toGroup(grp)
	}

	fields := make(render.Fields, 0, len(r.Header()))
	fields = append(fields, res, group)
	return render.Row{
		ID:     res,
		Fields: append(fields, verbs...),
	}
}

func asVerbs(verbs ...string) []string {
	const (
		verbLen    = 4
		unknownLen = 30
	)

	r := make([]string, 0, len(k8sVerbs)+1)
	for _, v := range k8sVerbs {
		r = append(r, toVerbIcon(hasVerb(verbs, v)))
	}

	var unknowns []string
	for _, v := range verbs {
		if hv, ok := httpTok8sVerbs[v]; ok {
			v = hv
		}
		if !hasVerb(k8sVerbs, v) && v != clusterWide {
			unknowns = append(unknowns, v)
		}
	}

	return append(r, resource.Truncate(strings.Join(unknowns, ","), unknownLen))
}

func toVerbIcon(ok bool) string {
	if ok {
		return "[green::b] ✓ [::]"
	}
	return "[orangered::b] 𐄂 [::]"
}

func hasVerb(verbs []string, verb string) bool {
	if len(verbs) == 1 && verbs[0] == clusterWide {
		return true
	}

	for _, v := range verbs {
		if hv, ok := httpTok8sVerbs[v]; ok {
			if hv == verb {
				return true
			}
		}
		if v == verb {
			return true
		}
	}

	return false
}

func toGroup(g string) string {
	if g == "" {
		return "v1"
	}
	return g
}

func showRoleBinding(app *App, _, resource, selection string) {
	ns, n := namespaced(selection)
	rb, err := app.Conn().DialOrDie().RbacV1().RoleBindings(ns).Get(n, metav1.GetOptions{})
	if err != nil {
		app.Flash().Errf("Unable to retrieve rolebindings for %s", selection)
		return
	}
	app.inject(NewRbac(fqn(ns, rb.RoleRef.Name), Role, selection))
}

func showClusterRoleBinding(app *App, ns, resource, selection string) {
	o, err := app.factory.Get("-", "rbac.authorization.k8s.io/v1/clusterrolebindings", selection, labels.Everything())
	if err != nil {
		app.Flash().Err(err)
		return
	}

	var crb rbacv1.ClusterRoleBinding
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &crb)
	if err != nil {
		app.Flash().Errf("Unable to retrieve clusterrolebindings for %s", selection)
		return
	}

	// BOZO!! Must make sure cluster roles are in cache prior to loading rbac view.
	app.factory.ForResource("-", "rbac.authorization.k8s.io/v1/clusterroles")
	app.factory.WaitForCacheSync()

	app.inject(NewRbac(crb.RoleRef.Name, ClusterRole, selection))
}

func showRBAC(app *App, ns, resource, selection string) {
	kind := ClusterRole
	if resource == "role" {
		kind = Role
	}
	app.inject(NewRbac(selection, kind, selection))
}
