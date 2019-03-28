package views

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
)

var fuHeader = append(resource.Row{"NAMESPACE", "NAME", "GROUP", "BINDING"}, rbacHeaderVerbs...)

type fuView struct {
	*tableView

	current     igniter
	cancel      context.CancelFunc
	subjectKind string
	subjectName string
	cache       resource.RowEvents
}

func newFuView(app *appView, subject, name string) *fuView {
	v := fuView{}
	{
		v.subjectKind, v.subjectName = v.mapSubject(subject), name
		v.tableView = newTableView(app, v.getTitle())
		v.colorerFn = rbacColorer
		v.current = app.content.GetPrimitive("main").(igniter)
		v.bindKeys()
	}

	return &v
}

// Init the view.
func (v *fuView) init(_ context.Context, ns string) {
	v.sortCol = sortColumn{1, len(rbacHeader), true}

	ctx, cancel := context.WithCancel(context.Background())
	v.cancel = cancel
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				log.Debug().Msg("FU Watch bailing out!")
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

func (v *fuView) bindKeys() {
	delete(v.actions, KeyShiftA)

	v.actions[tcell.KeyEscape] = newKeyAction("Reset", v.resetCmd, false)
	v.actions[KeySlash] = newKeyAction("Filter", v.activateCmd, false)
	v.actions[KeyP] = newKeyAction("Previous", v.app.prevCmd, false)

	v.actions[KeyShiftS] = newKeyAction("Sort Namespace", v.sortColCmd(0), true)
	v.actions[KeyShiftN] = newKeyAction("Sort Name", v.sortColCmd(1), true)
	v.actions[KeyShiftO] = newKeyAction("Sort Group", v.sortColCmd(2), true)
	v.actions[KeyShiftB] = newKeyAction("Sort Binding", v.sortColCmd(3), true)
}

func (v *fuView) getTitle() string {
	return fmt.Sprintf(rbacTitleFmt, "Fu", v.subjectKind+":"+v.subjectName)
}

func (v *fuView) refresh() {
	data, err := v.reconcile()
	if err != nil {
		log.Error().Err(err).Msgf("Unable to reconcile for %s:%s", v.subjectKind, v.subjectName)
	}
	v.update(data)
}

func (v *fuView) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.cmdBuff.empty() {
		v.cmdBuff.reset()
		return nil
	}

	return v.backCmd(evt)
}

func (v *fuView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
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

func (v *fuView) hints() hints {
	return v.actions.toHints()
}

func (v *fuView) reconcile() (resource.TableData, error) {
	evts, errs := v.clusterPolicies()
	if len(errs) > 0 {
		for _, err := range errs {
			log.Debug().Err(err).Msg("Unable to find cluster policies")
		}
		return resource.TableData{}, errs[0]
	}

	nevts, errs := v.namespacePolicies()
	if len(errs) > 0 {
		for _, err := range errs {
			log.Debug().Err(err).Msg("Unable to find cluster policies")
		}
		return resource.TableData{}, errs[0]
	}

	for k, v := range nevts {
		evts[k] = v
	}

	data := resource.TableData{
		Header:    fuHeader,
		Rows:      make(resource.RowEvents, len(evts)),
		Namespace: "*",
	}

	noDeltas := make(resource.Row, len(fuHeader))
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

func (v *fuView) clusterPolicies() (resource.RowEvents, []error) {
	var errs []error
	evts := make(resource.RowEvents)

	crbs, err := v.app.conn().DialOrDie().Rbac().ClusterRoleBindings().List(metav1.ListOptions{})
	if err != nil {
		return evts, errs
	}

	var roles []string
	for _, c := range crbs.Items {
		for _, s := range c.Subjects {
			if s.Kind == v.subjectKind && s.Name == v.subjectName {
				roles = append(roles, c.RoleRef.Name)
			}
		}
	}

	for _, r := range roles {
		cr, err := v.app.conn().DialOrDie().Rbac().ClusterRoles().Get(r, metav1.GetOptions{})
		if err != nil {
			errs = append(errs, err)
		}
		e := v.parseRules("*", r, cr.Rules)
		for k, v := range e {
			evts[k] = v
		}
	}

	return evts, errs
}

func (v *fuView) namespacePolicies() (resource.RowEvents, []error) {
	var errs []error
	evts := make(resource.RowEvents)

	rbs, err := v.app.conn().DialOrDie().Rbac().RoleBindings("").List(metav1.ListOptions{})
	if err != nil {
		return evts, errs
	}

	type nsRole struct {
		ns, role string
	}
	var roles []nsRole
	for _, rb := range rbs.Items {
		for _, s := range rb.Subjects {
			if s.Kind == v.subjectKind && s.Name == v.subjectName {
				roles = append(roles, nsRole{rb.Namespace, rb.RoleRef.Name})
			}
		}
	}

	for _, r := range roles {
		cr, err := v.app.conn().DialOrDie().Rbac().Roles(r.ns).Get(r.role, metav1.GetOptions{})
		if err != nil {
			errs = append(errs, err)
		}
		e := v.parseRules(r.ns, r.role, cr.Rules)
		for k, v := range e {
			evts[k] = v
		}
	}

	return evts, errs
}

func (v *fuView) namespace(ns, n string) string {
	return ns + "/" + n
}

func (v *fuView) parseRules(ns, binding string, rules []rbacv1.PolicyRule) resource.RowEvents {
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
					m[v.namespace(ns, n)] = &resource.RowEvent{
						Fields: v.prepRow(ns, n, grp, binding, r.Verbs),
					}
				}
				m[v.namespace(ns, k)] = &resource.RowEvent{
					Fields: v.prepRow(ns, k, grp, binding, r.Verbs),
				}
			}
		}
		for _, nres := range r.NonResourceURLs {
			if nres[0] != '/' {
				nres = "/" + nres
			}
			m[v.namespace(ns, nres)] = &resource.RowEvent{
				Fields: v.prepRow(ns, nres, resource.NAValue, binding, r.Verbs),
			}
		}
	}

	return m
}

func (v *fuView) prepRow(ns, res, grp, binding string, verbs []string) resource.Row {
	const (
		nameLen  = 60
		groupLen = 30
		nsLen    = 30
	)

	if grp != resource.NAValue {
		grp = toGroup(grp)
	}

	return v.makeRow(ns, res, grp, binding, asVerbs(verbs...))
}

func (*fuView) makeRow(ns, res, group, binding string, verbs []string) resource.Row {
	r := make(resource.Row, 0, len(fuHeader))
	r = append(r, ns, res, group, binding)

	return append(r, verbs...)
}

func (v *fuView) mapSubject(subject string) string {
	switch subject {
	case "g":
		return "Group"
	case "s":
		return "ServiceAccount"
	default:
		return "User"
	}
}
