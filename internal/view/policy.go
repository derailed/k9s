package view

import (
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

const (
	policyTitle = "Policy"
	group       = "Group"
	user        = "User"
	sa          = "ServiceAccount"
	allVerbs    = "*"
)

type (
	namespacedRole struct {
		ns, role string
	}

	// Policy presents a RBAC policy viewer.
	Policy struct {
		ResourceViewer
	}
)

// NewPolicy returns a new viewer.
func NewPolicy(gvr client.GVR) *Policy {
	p := Policy{
		ResourceViewer: NewBrowser(gvr),
	}
	p.GetTable().SetColorerFn(render.Policy{}.ColorerFunc())
	p.SetBindKeysFn(p.bindKeys)
	p.GetTable().SetSortCol(1, len(render.Policy{}.Header(render.AllNamespaces)), false)

	return &p
}

func (p *Policy) Name() string {
	return "policy"
}

// func (p *Policy) Start() {
// 	p.Stop()
// 	ctx, cancel := context.WithCancel(context.Background())
// 	p.cancel = cancel
// 	go func(ctx context.Context) {
// 		for {
// 			select {
// 			case <-ctx.Done():
// 				return
// 			case <-time.After(time.Duration(p.app.Config.K9s.GetRefreshRate()) * time.Second):
// 				p.refresh()
// 			}
// 		}
// 	}(ctx)
// }

func (p *Policy) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyActions{
		// tcell.KeyEscape: ui.NewKeyAction("Back", p.resetCmd, false),
		// ui.KeySlash:     ui.NewKeyAction("Filter", p.activateCmd, false),
		ui.KeyShiftP: ui.NewKeyAction("Sort Namespace", p.GetTable().SortColCmd(0, true), false),
		ui.KeyShiftN: ui.NewKeyAction("Sort Name", p.GetTable().SortColCmd(1, true), false),
		ui.KeyShiftO: ui.NewKeyAction("Sort Group", p.GetTable().SortColCmd(2, true), false),
		ui.KeyShiftB: ui.NewKeyAction("Sort Binding", p.GetTable().SortColCmd(3, true), false),
	})
}

func (p *Policy) getTitle() string {
	// BOZO!!
	// return fmt.Sprintf(rbacTitleFmt, policyTitle, p.subjectKind+":"+p.subjectName, p.GetRowCount())
	return ""
}

// func (p *Policy) refresh() {
// 	log.Debug().Msgf(">>>>>>>>>>>>>>> Refreshing Policies")
// 	// BOZO!!
// 	defer func(t time.Time) {
// 		log.Debug().Msgf("Policy Refresh elapsed %v", time.Since(t))
// 	}(time.Now())

// 	data, err := p.reconcile()
// 	if err != nil {
// 		log.Error().Err(err).Msgf("Refresh for %s:%s", p.subjectKind, p.subjectName)
// 		p.app.Flash().Err(err)
// 	}
// 	p.app.QueueUpdateDraw(func() {
// 		p.Update(data)
// 	})
// }

// func (p *Policy) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
// 	if !p.GetTable().SearchBuff().Empty() {
// 		p.GetTable().SearchBuff().Reset()
// 		return nil
// 	}

// 	return p.backCmd(evt)
// }

// func (p *Policy) backCmd(evt *tcell.EventKey) *tcell.EventKey {
// 	if p.cancel != nil {
// 		p.cancel()
// 	}

// 	if p.SearchBuff().IsActive() {
// 		p.SearchBuff().Reset()
// 		return nil
// 	}

// 	return p.app.PrevCmd(evt)
// }

// func (p *Policy) reconcile() (render.TableData, error) {
// 	// BOZO!!
// 	defer func(t time.Time) {
// 		log.Debug().Msgf("Policy Reconcile elapsed %v", time.Since(t))
// 	}(time.Now())

// 	var table render.TableData

// 	evts, errs := p.fetchClusterRoleBindings()
// 	if len(errs) > 0 {
// 		for _, err := range errs {
// 			log.Error().Err(err).Msg("Unable to find cluster policies")
// 		}
// 		return table, errs[0]
// 	}

// 	nevts, errs := p.namespacedPolicies()
// 	if len(errs) > 0 {
// 		for _, err := range errs {
// 			log.Error().Err(err).Msg("Unable to find cluster policies")
// 		}
// 		return table, errs[0]
// 	}

// 	for _, v := range nevts {
// 		evts = append(evts, v)
// 	}

// 	return buildTable(p, evts), nil
// }

// Protocol...

// func (p *Policy) Header() render.HeaderRow {
// 	return render.Policy{}.Header(render.AllNamespaces)
// }

// func (p *Policy) GetCache() render.RowEvents {
// 	return p.cache
// }

// func (p *Policy) SetCache(evts render.RowEvents) {
// 	p.cache = evts
// }

// func (p *Policy) fetchClusterRoleBindings() (render.Rows, []error) {
// 	var errs []error
// 	oo, err := p.app.factory.List(render.ClusterScope, "rbac.authorization.k8s.io/v1/clusterrolebindings", labels.Everything())
// 	if err != nil {
// 		return nil, append(errs, err)
// 	}

// 	roles := make([]string, 0, len(oo))
// 	for _, o := range oo {
// 		var crb rbacv1.ClusterRoleBinding
// 		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &crb)
// 		if err != nil {
// 			errs = append(errs, err)
// 			continue
// 		}
// 		for _, s := range crb.Subjects {
// 			if s.Kind == p.subjectKind && s.Name == p.subjectName {
// 				roles = append(roles, crb.RoleRef.Name)
// 			}
// 		}
// 	}

// 	rows := make(render.Rows, 0, len(oo))
// 	for _, role := range roles {
// 		o, err := p.app.factory.Get(render.ClusterScope, "rbac.authorization.k8s.io/v1/clusterroles", role, labels.Everything())
// 		if err != nil {
// 			return nil, append(errs, err)
// 		}
// 		var cr rbacv1.ClusterRole
// 		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cr)
// 		if err != nil {
// 			errs = append(errs, err)
// 			continue
// 		}

// 		for _, v := range p.parseRules("*", "CR:"+role, cr.Rules) {
// 			rows = append(rows, v)
// 		}
// 	}

// 	return rows, errs
// }

// func (p *Policy) fetchRoleBindings() ([]namespacedRole, error) {
// 	oo, err := p.app.factory.List(render.AllNamespaces, "rbac.authorization.k8s.io/v1/rolebindings", labels.Everything())
// 	if err != nil {
// 		return nil, err
// 	}

// 	rr := make([]namespacedRole, 0, len(oo))
// 	for _, o := range oo {
// 		var rb rbacv1.RoleBinding
// 		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &rb)
// 		if err != nil {
// 			return nil, err
// 		}
// 		for _, s := range rb.Subjects {
// 			if s.Kind == p.subjectKind && s.Name == p.subjectName {
// 				rr = append(rr, namespacedRole{rb.Namespace, rb.RoleRef.Name})
// 			}
// 		}
// 	}

// 	return rr, nil
// }

// func (p *Policy) fetchClusterRoles(errs []error, rr []namespacedRole) (render.Rows, []error) {
// 	rows := make(render.Rows, 0, len(rr))
// 	for _, r := range rr {
// 		o, err := p.app.factory.Get(r.ns, "rbac.authorization.k8s.io/v1/clusterroles", r.role, labels.Everything())
// 		if err != nil {
// 			return nil, append(errs, err)
// 		}

// 		var cr rbacv1.ClusterRole
// 		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cr)
// 		if err != nil {
// 			errs = append(errs, err)
// 			continue
// 		}
// 		rows = append(rows, p.parseRules(r.ns, "RO:"+r.role, cr.Rules)...)
// 	}

// 	return rows, errs
// }

// func (p *Policy) namespacedPolicies() (render.Rows, []error) {
// 	var errs []error
// 	roles, err := p.fetchRoleBindings()
// 	if err != nil {
// 		errs = append(errs, err)
// 	}

// 	return p.fetchClusterRoles(errs, roles)
// }

// func (p *Policy) parseRules(ns, binding string, rules []rbacv1.PolicyRule) render.Rows {
// 	m := make(render.Rows, 0, len(rules))
// 	for _, r := range rules {
// 		for _, grp := range r.APIGroups {
// 			for _, res := range r.Resources {
// 				k := res
// 				if grp != "" {
// 					k = res + "." + grp
// 				}
// 				for _, na := range r.ResourceNames {
// 					n := fqn(k, na)
// 					m = append(m, render.Row{
// 						ID:     fqn(ns, n),
// 						Fields: append(policyRow(ns, n, grp, binding), asVerbs(r.Verbs)...),
// 					})
// 				}
// 				m = append(m, render.Row{
// 					ID:     fqn(ns, k),
// 					Fields: append(policyRow(ns, k, grp, binding), asVerbs(r.Verbs)...),
// 				})
// 			}
// 		}
// 		for _, nres := range r.NonResourceURLs {
// 			if nres[0] != '/' {
// 				nres = "/" + nres
// 			}
// 			m = append(m, render.Row{
// 				ID:     fqn(ns, nres),
// 				Fields: append(policyRow(ns, nres, "", binding), asVerbs(r.Verbs)...),
// 			})
// 		}
// 	}

// 	return m
// }

func policyRow(ns, res, grp, binding string) render.Fields {
	if grp != "" {
		grp = toGroup(grp)
	}

	r := make(render.Fields, 0, len(render.Policy{}.Header(render.AllNamespaces)))
	return append(r, ns, res, grp, binding)
}

func mapSubject(subject string) string {
	switch subject {
	case "g":
		return group
	case "s":
		return sa
	default:
		return user
	}
}

// func showSAPolicy(app *App, _, _, selection string) {
// 	_, n := client.Namespaced(selection)
// 	subject, err := mapFuSubject("ServiceAccount")
// 	if err != nil {
// 		app.Flash().Err(err)
// 		return
// 	}
// 	app.inject(NewPolicy(app, subject, n))
// }

func toGroup(g string) string {
	if g == "" {
		return "v1"
	}
	return g
}

func hasVerb(verbs []string, verb string) bool {
	if len(verbs) == 1 && verbs[0] == allVerbs {
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

func toVerbIcon(ok bool) string {
	if ok {
		return "[green::b] ✓ [::]"
	}
	return "[orangered::b] 𐄂 [::]"
}

func asVerbs(verbs []string) []string {
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
		if !hasVerb(k8sVerbs, v) && v != allVerbs {
			unknowns = append(unknowns, v)
		}
	}

	return append(r, render.Truncate(strings.Join(unknowns, ","), unknownLen))
}
