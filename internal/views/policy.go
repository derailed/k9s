package views

import (
	"context"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const policyTitle = "Policy"

var policyHeader = append(resource.Row{"NAMESPACE", "NAME", "API GROUP", "BINDING"}, rbacHeaderVerbs...)

type (
	namespacedRole struct {
		ns, role string
	}

	policyView struct {
		*tableView

		current     igniter
		cancel      context.CancelFunc
		subjectKind string
		subjectName string
		cache       resource.RowEvents
	}
)

func newPolicyView(app *appView, subject, name string) *policyView {
	v := policyView{}
	{
		v.subjectKind, v.subjectName = mapSubject(subject), name
		v.tableView = newTableView(app, v.getTitle())
		v.colorerFn = rbacColorer
		v.current = app.content.GetPrimitive("main").(igniter)
		v.bindKeys()
	}

	return &v
}

// Init the view.
func (v *policyView) init(c context.Context, ns string) {
	v.sortCol = sortColumn{1, len(rbacHeader), false}

	ctx, cancel := context.WithCancel(c)
	v.cancel = cancel
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
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

func (v *policyView) bindKeys() {
	delete(v.actions, KeyShiftA)

	v.actions[tcell.KeyEscape] = newKeyAction("Reset", v.resetCmd, false)
	v.actions[KeySlash] = newKeyAction("Filter", v.activateCmd, false)
	v.actions[KeyP] = newKeyAction("Previous", v.app.prevCmd, false)

	v.actions[KeyShiftS] = newKeyAction("Sort Namespace", v.sortColCmd(0), true)
	v.actions[KeyShiftN] = newKeyAction("Sort Name", v.sortColCmd(1), true)
	v.actions[KeyShiftO] = newKeyAction("Sort Group", v.sortColCmd(2), true)
	v.actions[KeyShiftB] = newKeyAction("Sort Binding", v.sortColCmd(3), true)
}

func (v *policyView) getTitle() string {
	return fmt.Sprintf(rbacTitleFmt, policyTitle, v.subjectKind+":"+v.subjectName)
}

func (v *policyView) refresh() {
	data, err := v.reconcile()
	if err != nil {
		log.Error().Err(err).Msgf("Unable to reconcile for %s:%s", v.subjectKind, v.subjectName)
	}
	v.update(data)
}

func (v *policyView) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.cmdBuff.empty() {
		v.cmdBuff.reset()
		return nil
	}

	return v.backCmd(evt)
}

func (v *policyView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cancel != nil {
		v.cancel()
	}

	if v.cmdBuff.isActive() {
		v.cmdBuff.reset()
		return nil
	}

	v.app.inject(v.current)

	return nil
}

func (v *policyView) hints() hints {
	return v.actions.toHints()
}

func (v *policyView) reconcile() (resource.TableData, error) {
	var table resource.TableData

	evts, errs := v.clusterPolicies()
	if len(errs) > 0 {
		for _, err := range errs {
			log.Debug().Err(err).Msg("Unable to find cluster policies")
		}
		return table, errs[0]
	}

	nevts, errs := v.namespacedPolicies()
	if len(errs) > 0 {
		for _, err := range errs {
			log.Debug().Err(err).Msg("Unable to find cluster policies")
		}
		return table, errs[0]
	}

	for k, v := range nevts {
		evts[k] = v
	}

	return buildTable(v, evts), nil
}

// Protocol...

func (v *policyView) header() resource.Row {
	return policyHeader
}

func (v *policyView) getCache() resource.RowEvents {
	return v.cache
}

func (v *policyView) setCache(evts resource.RowEvents) {
	v.cache = evts
}

func (v *policyView) clusterPolicies() (resource.RowEvents, []error) {
	var errs []error
	evts := make(resource.RowEvents)

	crbs, err := v.app.conn().DialOrDie().Rbac().ClusterRoleBindings().List(metav1.ListOptions{})
	if err != nil {
		return evts, errs
	}

	var rr []string
	for _, crb := range crbs.Items {
		for _, s := range crb.Subjects {
			if s.Kind == v.subjectKind && s.Name == v.subjectName {
				rr = append(rr, crb.RoleRef.Name)
			}
		}
	}

	for _, r := range rr {
		role, err := v.app.conn().DialOrDie().Rbac().ClusterRoles().Get(r, metav1.GetOptions{})
		if err != nil {
			errs = append(errs, err)
		}
		for k, v := range v.parseRules("*", "CR:"+r, role.Rules) {
			evts[k] = v
		}
	}

	return evts, errs
}

func (v policyView) loadRoleBindings() ([]namespacedRole, error) {
	var rr []namespacedRole

	dial := v.app.conn().DialOrDie().Rbac()
	rbs, err := dial.RoleBindings("").List(metav1.ListOptions{})
	if err != nil {
		return rr, err
	}

	for _, rb := range rbs.Items {
		for _, s := range rb.Subjects {
			if s.Kind == v.subjectKind && s.Name == v.subjectName {
				rr = append(rr, namespacedRole{rb.Namespace, rb.RoleRef.Name})
			}
		}
	}

	return rr, nil
}

func (v *policyView) loadRoles(errs []error, rr []namespacedRole) (resource.RowEvents, []error) {
	var (
		dial = v.app.conn().DialOrDie().Rbac()
		evts = make(resource.RowEvents)
	)
	for _, r := range rr {
		if cr, err := dial.Roles(r.ns).Get(r.role, metav1.GetOptions{}); err != nil {
			errs = append(errs, err)
		} else {
			for k, v := range v.parseRules(r.ns, "RO:"+r.role, cr.Rules) {
				evts[k] = v
			}
		}
	}

	return evts, errs
}

func (v *policyView) namespacedPolicies() (resource.RowEvents, []error) {
	var errs []error
	rr, err := v.loadRoleBindings()
	if err != nil {
		errs = append(errs, err)
	}

	evts, errs := v.loadRoles(errs, rr)
	return evts, errs
}

func (v *policyView) parseRules(ns, binding string, rules []rbacv1.PolicyRule) resource.RowEvents {
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
					m[fqn(ns, n)] = &resource.RowEvent{
						Fields: append(policyRow(ns, n, grp, binding), r.Verbs...),
					}
				}
				m[fqn(ns, k)] = &resource.RowEvent{
					Fields: append(policyRow(ns, k, grp, binding), r.Verbs...),
				}
			}
		}
		for _, nres := range r.NonResourceURLs {
			if nres[0] != '/' {
				nres = "/" + nres
			}
			m[fqn(ns, nres)] = &resource.RowEvent{
				Fields: append(policyRow(ns, nres, resource.NAValue, binding), r.Verbs...),
			}
		}
	}

	return m
}

func policyRow(ns, res, grp, binding string) resource.Row {
	if grp != resource.NAValue {
		grp = toGroup(grp)
	}

	r := make(resource.Row, 0, len(policyHeader))
	return append(r, ns, res, grp, binding)
}

func mapSubject(subject string) string {
	switch subject {
	case "g":
		return "Group"
	case "s":
		return "ServiceAccount"
	default:
		return "User"
	}
}
