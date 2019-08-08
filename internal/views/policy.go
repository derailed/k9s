package views

import (
	"context"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
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

		current     ui.Igniter
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
		v.SetColorerFn(rbacColorer)
		v.current = app.Frame().GetPrimitive("main").(ui.Igniter)
		v.bindKeys()
	}

	return &v
}

// Init the view.
func (v *policyView) Init(c context.Context, ns string) {
	v.SetSortCol(1, len(rbacHeader), false)

	ctx, cancel := context.WithCancel(c)
	v.cancel = cancel
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(v.app.Config.K9s.GetRefreshRate()) * time.Second):
				v.refresh()
				v.app.Draw()
			}
		}
	}(ctx)

	v.refresh()
	v.SelectRow(1, true)
	v.app.SetFocus(v)
}

func (v *policyView) bindKeys() {
	v.RmAction(ui.KeyShiftA)

	v.SetActions(ui.KeyActions{
		tcell.KeyEscape: ui.NewKeyAction("Reset", v.resetCmd, false),
		ui.KeySlash:     ui.NewKeyAction("Filter", v.activateCmd, false),
		ui.KeyP:         ui.NewKeyAction("Previous", v.app.prevCmd, false),
		ui.KeyShiftS:    ui.NewKeyAction("Sort Namespace", v.SortColCmd(0), true),
		ui.KeyShiftN:    ui.NewKeyAction("Sort Name", v.SortColCmd(1), true),
		ui.KeyShiftO:    ui.NewKeyAction("Sort Group", v.SortColCmd(2), true),
		ui.KeyShiftB:    ui.NewKeyAction("Sort Binding", v.SortColCmd(3), true),
	})
}

func (v *policyView) getTitle() string {
	return fmt.Sprintf(rbacTitleFmt, policyTitle, v.subjectKind+":"+v.subjectName)
}

func (v *policyView) refresh() {
	data, err := v.reconcile()
	if err != nil {
		log.Error().Err(err).Msgf("Unable to reconcile for %s:%s", v.subjectKind, v.subjectName)
	}
	v.Update(data)
}

func (v *policyView) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.Cmd().Empty() {
		v.Cmd().Reset()
		return nil
	}

	return v.backCmd(evt)
}

func (v *policyView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cancel != nil {
		v.cancel()
	}

	if v.Cmd().IsActive() {
		v.Cmd().Reset()
		return nil
	}

	v.app.inject(v.current)

	return nil
}

func (v *policyView) Hints() ui.Hints {
	return v.Hints()
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

	crbs, err := v.app.Conn().DialOrDie().RbacV1().ClusterRoleBindings().List(metav1.ListOptions{})
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
		role, err := v.app.Conn().DialOrDie().RbacV1().ClusterRoles().Get(r, metav1.GetOptions{})
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

	dial := v.app.Conn().DialOrDie().RbacV1()
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
		dial = v.app.Conn().DialOrDie().RbacV1()
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
						Fields: append(policyRow(ns, n, grp, binding), asVerbs(r.Verbs...)...),
					}
				}
				m[fqn(ns, k)] = &resource.RowEvent{
					Fields: append(policyRow(ns, k, grp, binding), asVerbs(r.Verbs...)...),
				}
			}
		}
		for _, nres := range r.NonResourceURLs {
			if nres[0] != '/' {
				nres = "/" + nres
			}
			m[fqn(ns, nres)] = &resource.RowEvent{
				Fields: append(policyRow(ns, nres, resource.NAValue, binding), asVerbs(r.Verbs...)...),
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
