package views

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

type (
	roleKind = int8

	rbacView struct {
		*tableView

		app      *appView
		current  ui.Igniter
		cancel   context.CancelFunc
		roleType roleKind
		roleName string
		cache    resource.RowEvents
	}
)

func newRBACView(app *appView, ns, name string, kind roleKind) *rbacView {
	v := rbacView{
		app:      app,
		roleName: name,
		roleType: kind,
	}
	v.tableView = newTableView(app, v.getTitle())
	v.SetActiveNS(ns)
	v.SetColorerFn(rbacColorer)
	v.current = app.Frame().GetPrimitive("main").(ui.Igniter)
	v.bindKeys()

	return &v
}

// Init the view.
func (v *rbacView) Init(c context.Context, ns string) {
	v.SetSortCol(1, len(rbacHeader), true)

	ctx, cancel := context.WithCancel(c)
	v.cancel = cancel
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(v.app.Config.K9s.GetRefreshRate()) * time.Second):
				v.app.QueueUpdateDraw(func() {
					v.refresh()
				})
			}
		}
	}(ctx)

	v.refresh()
	v.app.SetHints(v.Hints())
	v.app.SetFocus(v)
}

func (v *rbacView) bindKeys() {
	v.RmAction(ui.KeyShiftA)

	v.SetActions(ui.KeyActions{
		tcell.KeyEscape: ui.NewKeyAction("Reset", v.resetCmd, false),
		ui.KeySlash:     ui.NewKeyAction("Filter", v.activateCmd, false),
		ui.KeyP:         ui.NewKeyAction("Previous", v.app.prevCmd, false),
		ui.KeyShiftO:    ui.NewKeyAction("Sort APIGroup", v.SortColCmd(1), false),
	})
}

func (v *rbacView) getTitle() string {
	return skinTitle(fmt.Sprintf(rbacTitleFmt, rbacTitle, v.roleName), v.app.Styles.Frame())
}

func (v *rbacView) refresh() {
	data, err := v.reconcile(v.ActiveNS(), v.roleName, v.roleType)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to reconcile for %s:%d", v.roleName, v.roleType)
	}
	v.Update(data)
}

func (v *rbacView) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.SearchBuff().Empty() {
		v.SearchBuff().Reset()
		return nil
	}

	return v.backCmd(evt)
}

func (v *rbacView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cancel != nil {
		v.cancel()
	}

	if v.SearchBuff().IsActive() {
		v.SearchBuff().Reset()
		return nil
	}

	v.app.inject(v.current)

	return nil
}

func (v *rbacView) reconcile(ns, name string, kind roleKind) (resource.TableData, error) {
	var table resource.TableData

	evts, err := v.rowEvents(ns, name, kind)
	if err != nil {
		return table, err
	}

	return buildTable(v, evts), nil
}

func (v *rbacView) header() resource.Row {
	return rbacHeader
}

func (v *rbacView) getCache() resource.RowEvents {
	return v.cache
}

func (v *rbacView) setCache(evts resource.RowEvents) {
	v.cache = evts
}

func (v *rbacView) rowEvents(ns, name string, kind roleKind) (resource.RowEvents, error) {
	var (
		evts resource.RowEvents
		err  error
	)

	switch kind {
	case clusterRole:
		evts, err = v.clusterPolicies(name)
	case role:
		evts, err = v.namespacedPolicies(name)
	default:
		return evts, fmt.Errorf("Expecting clusterrole/role but found %d", kind)
	}
	if err != nil {
		log.Error().Err(err).Msg("Unable to load CR")
		return evts, err
	}

	return evts, nil
}

func (v *rbacView) clusterPolicies(name string) (resource.RowEvents, error) {
	cr, err := v.app.Conn().DialOrDie().RbacV1().ClusterRoles().Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return v.parseRules(cr.Rules), nil
}

func (v *rbacView) namespacedPolicies(path string) (resource.RowEvents, error) {
	ns, na := namespaced(path)
	cr, err := v.app.Conn().DialOrDie().RbacV1().Roles(ns).Get(na, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return v.parseRules(cr.Rules), nil
}

func (v *rbacView) parseRules(rules []rbacv1.PolicyRule) resource.RowEvents {
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
