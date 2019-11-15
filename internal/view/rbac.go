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
	ClusterRole roleKind = iota
	Role

	all          = "*"
	rbacTitle    = "Rbac"
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

	httpTok8sVerbs = map[string]string{
		"post": "create",
		"put":  "update",
	}
)

type roleKind = int8

// Rbac presents an RBAC policy viewer.
type Rbac struct {
	*Table

	app      *App
	cancelFn context.CancelFunc
	roleType roleKind
	roleName string
	cache    resource.RowEvents
}

// NewRbac returns a new viewer.
func NewRbac(app *App, ns, name string, kind roleKind) *Rbac {
	r := Rbac{
		app:      app,
		roleName: name,
		roleType: kind,
	}
	r.Table = NewTable(r.getTitle())

	return &r
}

// Init initializes the view.
func (r *Rbac) Init(ctx context.Context) {
	r.ActiveNS = r.app.Config.ActiveNamespace()
	r.SetColorerFn(rbacColorer)
	r.Table.Init(ctx)
	r.bindKeys()

	r.Start()
	r.SetSortCol(1, len(rbacHeader), true)
	r.refresh()
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

// Stop terminates the viewer updater.
func (r *Rbac) Stop() {
	if r.cancelFn != nil {
		r.cancelFn()
	}
}

// Name returns the component name.
func (r *Rbac) Name() string {
	return rbacTitle
}

func (r *Rbac) bindKeys() {
	r.RmAction(ui.KeyShiftA)
	r.RmAction(tcell.KeyCtrlSpace)
	r.RmAction(ui.KeySpace)

	r.AddActions(ui.KeyActions{
		tcell.KeyEscape: ui.NewKeyAction("Reset", r.resetCmd, false),
		ui.KeySlash:     ui.NewKeyAction("Filter", r.activateCmd, false),
		ui.KeyShiftO:    ui.NewKeyAction("Sort APIGroup", r.SortColCmd(1), false),
	})
}

func (r *Rbac) getTitle() string {
	return skinTitle(fmt.Sprintf(rbacTitleFmt, rbacTitle, r.roleName), r.app.Styles.Frame())
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
}

func (r *Rbac) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.SearchBuff().Empty() {
		r.SearchBuff().Reset()
		return nil
	}

	return r.backCmd(evt)
}

func (r *Rbac) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	log.Debug().Msgf("!!!!RBAC back!!!")
	if r.cancelFn != nil {
		r.cancelFn()
	}

	if r.SearchBuff().IsActive() {
		r.SearchBuff().Reset()
		return nil
	}

	return r.app.PrevCmd(evt)
}

func (r *Rbac) reconcile(name string, kind roleKind) (resource.TableData, error) {
	var table resource.TableData

	evts, err := r.rowEvents(name, kind)
	if err != nil {
		return table, err
	}

	return buildTable(r, evts), nil
}

func (r *Rbac) header() resource.Row {
	return rbacHeader
}

func (r *Rbac) getCache() resource.RowEvents {
	return r.cache
}

func (r *Rbac) setCache(evts resource.RowEvents) {
	r.cache = evts
}

func (r *Rbac) rowEvents(name string, kind roleKind) (resource.RowEvents, error) {
	var (
		evts resource.RowEvents
		err  error
	)

	switch kind {
	case ClusterRole:
		evts, err = r.clusterPolicies(name)
	case Role:
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

func (r *Rbac) clusterPolicies(name string) (resource.RowEvents, error) {
	cr, err := r.app.Conn().DialOrDie().RbacV1().ClusterRoles().Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return r.parseRules(cr.Rules), nil
}

func (r *Rbac) namespacedPolicies(path string) (resource.RowEvents, error) {
	ns, na := namespaced(path)
	cr, err := r.app.Conn().DialOrDie().RbacV1().Roles(ns).Get(na, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return r.parseRules(cr.Rules), nil
}

func (r *Rbac) parseRules(rules []rbacv1.PolicyRule) resource.RowEvents {
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
