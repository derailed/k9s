package view

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

	// Policy presents a RBAC policy viewer.
	Policy struct {
		*Table

		cancel      context.CancelFunc
		subjectKind string
		subjectName string
		cache       resource.RowEvents
	}
)

// NewPolicy returns a new viewer.
func NewPolicy(app *App, subject, name string) *Policy {
	p := Policy{}
	p.subjectKind, p.subjectName = mapSubject(subject), name
	p.Table = NewTable(p.getTitle())
	p.SetColorerFn(rbacColorer)
	p.bindKeys()

	return &p
}

// Init the view.
func (p *Policy) Init(ctx context.Context) {
	p.Table.Init(ctx)

	p.SetSortCol(1, len(rbacHeader), false)
	p.Start()
	p.refresh()
	p.SelectRow(1, true)
}

func (p *Policy) Name() string {
	return "policy"
}

func (p *Policy) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel
	go func(ctx context.Context) {
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(time.Duration(p.app.Config.K9s.GetRefreshRate()) * time.Second):
				p.refresh()
				p.app.Draw()
			}
		}
	}(ctx)
}

func (p *Policy) Stop() {
	if p.cancel != nil {
		p.cancel()
	}
}

func (p *Policy) bindKeys() {
	p.RmAction(ui.KeyShiftA)

	p.AddActions(ui.KeyActions{
		tcell.KeyEscape: ui.NewKeyAction("Back", p.resetCmd, false),
		ui.KeySlash:     ui.NewKeyAction("Filter", p.activateCmd, false),
		ui.KeyShiftS:    ui.NewKeyAction("Sort Namespace", p.SortColCmd(0), false),
		ui.KeyShiftN:    ui.NewKeyAction("Sort Name", p.SortColCmd(1), false),
		ui.KeyShiftO:    ui.NewKeyAction("Sort Group", p.SortColCmd(2), false),
		ui.KeyShiftB:    ui.NewKeyAction("Sort Binding", p.SortColCmd(3), false),
	})
}

func (p *Policy) getTitle() string {
	return fmt.Sprintf(rbacTitleFmt, policyTitle, p.subjectKind+":"+p.subjectName)
}

func (p *Policy) refresh() {
	data, err := p.reconcile()
	if err != nil {
		log.Error().Err(err).Msgf("Refresh for %s:%s", p.subjectKind, p.subjectName)
		p.app.Flash().Err(err)
	}
	p.Update(data)
}

func (p *Policy) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !p.SearchBuff().Empty() {
		p.SearchBuff().Reset()
		return nil
	}

	return p.backCmd(evt)
}

func (p *Policy) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	if p.cancel != nil {
		p.cancel()
	}

	if p.SearchBuff().IsActive() {
		p.SearchBuff().Reset()
		return nil
	}

	return p.app.PrevCmd(evt)
}

func (p *Policy) reconcile() (resource.TableData, error) {
	var table resource.TableData

	evts, errs := p.clusterPolicies()
	if len(errs) > 0 {
		for _, err := range errs {
			log.Error().Err(err).Msg("Unable to find cluster policies")
		}
		return table, errs[0]
	}

	nevts, errs := p.namespacedPolicies()
	if len(errs) > 0 {
		for _, err := range errs {
			log.Error().Err(err).Msg("Unable to find cluster policies")
		}
		return table, errs[0]
	}

	for k, v := range nevts {
		evts[k] = v
	}

	return buildTable(p, evts), nil
}

// Protocol...

func (p *Policy) header() resource.Row {
	return policyHeader
}

func (p *Policy) getCache() resource.RowEvents {
	return p.cache
}

func (p *Policy) setCache(evts resource.RowEvents) {
	p.cache = evts
}

func (p *Policy) clusterPolicies() (resource.RowEvents, []error) {
	var errs []error
	evts := make(resource.RowEvents)

	crbs, err := p.app.Conn().DialOrDie().RbacV1().ClusterRoleBindings().List(metav1.ListOptions{})
	if err != nil {
		return evts, errs
	}

	var rr []string
	for _, crb := range crbs.Items {
		for _, s := range crb.Subjects {
			if s.Kind == p.subjectKind && s.Name == p.subjectName {
				rr = append(rr, crb.RoleRef.Name)
			}
		}
	}

	for _, r := range rr {
		role, err := p.app.Conn().DialOrDie().RbacV1().ClusterRoles().Get(r, metav1.GetOptions{})
		if err != nil {
			errs = append(errs, err)
		}
		for k, v := range p.parseRules("*", "CR:"+r, role.Rules) {
			evts[k] = v
		}
	}

	return evts, errs
}

func (p *Policy) loadRoleBindings() ([]namespacedRole, error) {
	var rr []namespacedRole

	dial := p.app.Conn().DialOrDie().RbacV1()
	rbs, err := dial.RoleBindings("").List(metav1.ListOptions{})
	if err != nil {
		return rr, err
	}

	for _, rb := range rbs.Items {
		for _, s := range rb.Subjects {
			if s.Kind == p.subjectKind && s.Name == p.subjectName {
				rr = append(rr, namespacedRole{rb.Namespace, rb.RoleRef.Name})
			}
		}
	}

	return rr, nil
}

func (p *Policy) loadRoles(errs []error, rr []namespacedRole) (resource.RowEvents, []error) {
	var (
		dial = p.app.Conn().DialOrDie().RbacV1()
		evts = make(resource.RowEvents)
	)
	for _, r := range rr {
		if cr, err := dial.Roles(r.ns).Get(r.role, metav1.GetOptions{}); err != nil {
			errs = append(errs, err)
		} else {
			for k, v := range p.parseRules(r.ns, "RO:"+r.role, cr.Rules) {
				evts[k] = v
			}
		}
	}

	return evts, errs
}

func (p *Policy) namespacedPolicies() (resource.RowEvents, []error) {
	var errs []error
	rr, err := p.loadRoleBindings()
	if err != nil {
		errs = append(errs, err)
	}

	evts, errs := p.loadRoles(errs, rr)
	return evts, errs
}

func (p *Policy) parseRules(ns, binding string, rules []rbacv1.PolicyRule) resource.RowEvents {
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
