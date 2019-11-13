package view

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	clusterRole roleKind = iota
	role

	all          = "*"
	rbacTitle    = "RBAC"
	rbacTitleFmt = " [fg:bg:b]%s([hilite:bg:b]%s[fg:bg:-])"
)

var (
	rbacHeaderVerbs = resource.Row{
		"GET   ",
		"LIST  ",
		"DLIST ",
		"WATCH ",
		"CREATE",
		"PATCH ",
		"UPDATE",
		"DELETE",
		"EXTRAS",
	}

	rbacHeader = append(resource.Row{"NAME", "API GROUP"}, rbacHeaderVerbs...)

	k8sVerbs = []string{
		"get",
		"list",
		"deletecollection",
		"watch",
		"create",
		"patch",
		"update",
		"delete",
	}

	httpVerbs = []string{
		"get",
		"post",
		"put",
		"patch",
		"delete",
		"options",
	}

	httpTok8sVerbs = map[string]string{
		"post": "create",
		"put":  "update",
	}
)

type roleKind = int8

// RBAC presents an RBAC policy viewer.
type RBAC struct {
	*Table

	app      *App
	cancelFn context.CancelFunc
	roleType roleKind
	roleName string
	cache    resource.RowEvents
}

// NewRBAC returns a new viewer.
func NewRBAC(app *App, ns, name string, kind roleKind) *RBAC {
	r := RBAC{
		app:      app,
		roleName: name,
		roleType: kind,
	}
	r.Table = NewTable(r.getTitle())
	r.SetActiveNS(ns)
	r.SetColorerFn(rbacColorer)
	r.bindKeys()

	return &r
}

// Init initializes the view.
func (r *RBAC) Init(ctx context.Context) {
	r.Table.Init(ctx)

	r.Start()
	r.SetSortCol(1, len(rbacHeader), true)
	r.refresh()
}

// Start watches for viewer updates
func (r *RBAC) Start() {
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

// Stop terminates the viewer updater.
func (r *RBAC) Stop() {
	if r.cancelFn != nil {
		r.cancelFn()
	}
}

// Name returns the component name.
func (r *RBAC) Name() string {
	return rbacTitle
}

func (r *RBAC) bindKeys() {
	r.RmAction(ui.KeyShiftA)

	r.AddActions(ui.KeyActions{
		tcell.KeyEscape: ui.NewKeyAction("Reset", r.resetCmd, false),
		ui.KeySlash:     ui.NewKeyAction("Filter", r.activateCmd, false),
		ui.KeyP:         ui.NewKeyAction("Previous", r.app.PrevCmd, false),
		ui.KeyShiftO:    ui.NewKeyAction("Sort APIGroup", r.SortColCmd(1), false),
	})
}

func (r *RBAC) getTitle() string {
	return skinTitle(fmt.Sprintf(rbacTitleFmt, rbacTitle, r.roleName), r.app.Styles.Frame())
}

func (r *RBAC) refresh() {
	data, err := r.reconcile(r.ActiveNS(), r.roleName, r.roleType)
	if err != nil {
		log.Error().Err(err).Msgf("Refresh for %s:%d", r.roleName, r.roleType)
		r.app.Flash().Err(err)
	}
	r.Update(data)
}

func (r *RBAC) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.SearchBuff().Empty() {
		r.SearchBuff().Reset()
		return nil
	}

	return r.backCmd(evt)
}

func (r *RBAC) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if r.cancelFn != nil {
		r.cancelFn()
	}

	if r.SearchBuff().IsActive() {
		r.SearchBuff().Reset()
		return nil
	}

	return r.app.PrevCmd(evt)
}

func (r *RBAC) reconcile(ns, name string, kind roleKind) (resource.TableData, error) {
	var table resource.TableData

	evts, err := r.rowEvents(ns, name, kind)
	if err != nil {
		return table, err
	}

	return buildTable(r, evts), nil
}

func (r *RBAC) header() resource.Row {
	return rbacHeader
}

func (r *RBAC) getCache() resource.RowEvents {
	return r.cache
}

func (r *RBAC) setCache(evts resource.RowEvents) {
	r.cache = evts
}

func (r *RBAC) rowEvents(ns, name string, kind roleKind) (resource.RowEvents, error) {
	var (
		evts resource.RowEvents
		err  error
	)

	switch kind {
	case clusterRole:
		evts, err = r.clusterPolicies(name)
	case role:
		evts, err = r.namespacedPolicies(name)
	default:
		return evts, fmt.Errorf("Expecting clusterrole/role but found %d", kind)
	}
	if err != nil {
		log.Error().Err(err).Msg("Unable to load CR")
		return evts, err
	}

	return evts, nil
}

func (r *RBAC) clusterPolicies(name string) (resource.RowEvents, error) {
	cr, err := r.app.Conn().DialOrDie().RbacV1().ClusterRoles().Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return r.parseRules(cr.Rules), nil
}

func (r *RBAC) namespacedPolicies(path string) (resource.RowEvents, error) {
	ns, na := namespaced(path)
	cr, err := r.app.Conn().DialOrDie().RbacV1().Roles(ns).Get(na, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return r.parseRules(cr.Rules), nil
}

func (r *RBAC) parseRules(rules []rbacv1.PolicyRule) resource.RowEvents {
	m := make(resource.RowEvents, len(rules))
	for _, r := range rules {
		for _, grp := range r.APIGroups {
			for _, res := range r.Resources {
				k := res
				if grp != "" {
					k = res + "." + grp
				}
				for _, na := range r.ResourceNames {
					n := fqn(k, na)
					m[n] = &resource.RowEvent{
						Fields: prepRow(n, grp, r.Verbs),
					}
				}
				m[k] = &resource.RowEvent{
					Fields: prepRow(k, grp, r.Verbs),
				}
			}
		}
		for _, nres := range r.NonResourceURLs {
			if nres[0] != '/' {
				nres = "/" + nres
			}
			m[nres] = &resource.RowEvent{
				Fields: prepRow(nres, resource.NAValue, r.Verbs),
			}
		}
	}

	return m
}

func prepRow(res, grp string, verbs []string) resource.Row {
	const (
		nameLen  = 60
		groupLen = 30
	)

	if grp != resource.NAValue {
		grp = toGroup(grp)
	}

	return makeRow(res, grp, asVerbs(verbs...))
}

func makeRow(res, group string, verbs []string) resource.Row {
	r := make(resource.Row, 0, len(rbacHeader))
	r = append(r, res, group)

	return append(r, verbs...)
}

func asVerbs(verbs ...string) resource.Row {
	const (
		verbLen    = 4
		unknownLen = 30
	)

	r := make(resource.Row, 0, len(k8sVerbs)+1)
	for _, v := range k8sVerbs {
		r = append(r, toVerbIcon(hasVerb(verbs, v)))
	}

	var unknowns []string
	for _, v := range verbs {
		if hv, ok := httpTok8sVerbs[v]; ok {
			v = hv
		}
		if !hasVerb(k8sVerbs, v) && v != all {
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
	if len(verbs) == 1 && verbs[0] == all {
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
