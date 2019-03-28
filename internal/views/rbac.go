package views

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	clusterRole roleKind = iota
	role

	all          = "*"
	rbacTitle    = "RBAC"
	rbacTitleFmt = " [aqua::b]%s([fuchsia::b]%s[aqua::-])"
)

type (
	roleKind = int8

	rbacView struct {
		*tableView

		current  igniter
		cancel   context.CancelFunc
		roleType roleKind
		roleName string
		cache    resource.RowEvents
	}
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
	rbacHeader = append(resource.Row{"NAME", "GROUP"}, rbacHeaderVerbs...)

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

func newRBACView(app *appView, ns, name string, kind roleKind) *rbacView {
	v := rbacView{}
	{
		v.roleName, v.roleType = name, kind
		v.tableView = newTableView(app, v.getTitle())
		v.currentNS = ns
		v.colorerFn = rbacColorer
		v.current = app.content.GetPrimitive("main").(igniter)
		v.bindKeys()
	}

	return &v
}

// Init the view.
func (v *rbacView) init(_ context.Context, ns string) {
	v.sortCol = sortColumn{1, len(rbacHeader), true}

	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("RBAC Watch bailing out!")
				return
			case <-time.After(time.Duration(v.app.config.K9s.RefreshRate) * time.Second):
				v.refresh()
				v.app.Draw()
			}
		}
	}(ctx)

	v.refresh()
	v.app.SetFocus(v)
}

func (v *rbacView) bindKeys() {
	delete(v.actions, KeyShiftA)

	v.actions[tcell.KeyEscape] = newKeyAction("Reset", v.resetCmd, false)
	v.actions[KeySlash] = newKeyAction("Filter", v.activateCmd, false)
	v.actions[KeyP] = newKeyAction("Previous", v.app.prevCmd, false)

	v.actions[KeyShiftO] = newKeyAction("Sort Groups", v.sortColCmd(1), true)
}

func (v *rbacView) getTitle() string {
	title := "ClusterRole"
	if v.roleType == role {
		title = "Role"
	}

	return fmt.Sprintf(rbacTitleFmt, title, v.roleName)
}

func (v *rbacView) hints() hints {
	return v.actions.toHints()
}

func (v *rbacView) refresh() {
	data, err := v.reconcile(v.currentNS, v.roleName, v.roleType)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to reconcile for %s:%d", v.roleName, v.roleType)
	}
	v.update(data)
}

func (v *rbacView) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.cmdBuff.empty() {
		v.cmdBuff.reset()
		return nil
	}

	return v.backCmd(evt)
}

func (v *rbacView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cancel != nil {
		v.cancel()
	}

	if v.cmdBuff.isActive() {
		v.cmdBuff.reset()
	} else {
		v.app.prevCmd(evt)
	}

	return nil
}

func (v *rbacView) reconcile(ns, name string, kind roleKind) (resource.TableData, error) {
	evts, err := v.rowEvents(ns, name, kind)
	if err != nil {
		return resource.TableData{}, err
	}

	data := resource.TableData{
		Header:    rbacHeader,
		Rows:      make(resource.RowEvents, len(evts)),
		Namespace: resource.NotNamespaced,
	}

	noDeltas := make(resource.Row, len(rbacHeader))
	if len(v.cache) == 0 {
		for k, ev := range evts {
			ev.Action = resource.New
			ev.Deltas = noDeltas
			data.Rows[k] = ev
		}
		v.cache = evts

		return data, nil
	}

	for k, ev := range evts {
		data.Rows[k] = ev

		newr := ev.Fields
		if _, ok := v.cache[k]; !ok {
			ev.Action, ev.Deltas = watch.Added, noDeltas
			continue
		}
		oldr := v.cache[k].Fields
		deltas := make(resource.Row, len(newr))
		if !reflect.DeepEqual(oldr, newr) {
			ev.Action = watch.Modified
			for i, field := range oldr {
				if field != newr[i] {
					deltas[i] = field
				}
			}
			ev.Deltas = deltas
		} else {
			ev.Action = resource.Unchanged
			ev.Deltas = noDeltas
		}
	}
	v.cache = evts

	for k := range v.cache {
		if _, ok := data.Rows[k]; !ok {
			delete(v.cache, k)
		}
	}

	return data, nil
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
	cr, err := v.app.conn().DialOrDie().Rbac().ClusterRoles().Get(name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	return v.parseRules(cr.Rules), nil
}

func (v *rbacView) namespacedPolicies(path string) (resource.RowEvents, error) {
	ns, na := namespaced(path)
	cr, err := v.app.conn().DialOrDie().Rbac().Roles(ns).Get(na, metav1.GetOptions{})
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
					n := k + "/" + na
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
