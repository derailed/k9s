package model

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

type Policy struct {
	Resource
}

func (p *Policy) List(ctx context.Context) ([]runtime.Object, error) {
	gvr, ok := ctx.Value(internal.KeyGVR).(string)
	if !ok {
		return nil, fmt.Errorf("expecting a context gvr")
	}
	kind, ok := ctx.Value(internal.KeySubjectKind).(string)
	if !ok {
		return nil, fmt.Errorf("expecting a context subject kind")
	}
	name, ok := ctx.Value(internal.KeySubjectName).(string)
	if !ok {
		return nil, fmt.Errorf("expecting a context subject name")
	}

	p.gvr = gvr
	crps, err := p.loadClusterRoleBinding(kind, name)
	if err != nil {
		return nil, err
	}
	rps, err := p.loadRoleBinding(kind, name)
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, 0, len(crps)+len(rps))
	for _, p := range crps {
		oo = append(oo, p)
	}
	for _, p := range rps {
		oo = append(oo, p)
	}

	return oo, nil
}

func (p *Policy) loadClusterRoleBinding(kind, name string) (render.Policies, error) {
	crbs, err := fetchClusterRoleBindings(p.factory)
	if err != nil {
		return nil, err
	}

	var nn []string
	for _, crb := range crbs {
		for _, s := range crb.Subjects {
			if s.Kind == kind && s.Name == name {
				nn = append(nn, crb.RoleRef.Name)
			}
		}
	}
	crs, err := p.fetchClusterRoles()
	if err != nil {
		return nil, err
	}

	rows := make(render.Policies, 0, len(nn))
	for _, cr := range crs {
		if !in(nn, cr.Name) {
			continue
		}
		rows = append(rows, parseRules("*", "CR:"+cr.Name, cr.Rules)...)
	}

	return rows, nil
}

func (p *Policy) loadRoleBinding(kind, name string) (render.Policies, error) {
	ss, err := p.fetchRoleBindingSubjects(kind, name)
	if err != nil {
		return nil, err
	}

	crs, err := p.fetchClusterRoles()
	if err != nil {
		return nil, err
	}
	rows := make(render.Policies, 0, len(crs))
	for _, cr := range crs {
		if !in(ss, "ClusterRole:"+cr.Name) {
			continue
		}
		rows = append(rows, parseRules("*", "CR:"+cr.Name, cr.Rules)...)
	}

	ros, err := p.fetchRoles()
	if err != nil {
		return nil, err
	}
	for _, ro := range ros {
		if !in(ss, "Role:"+ro.Name) {
			continue
		}
		log.Debug().Msgf("Loading rules for role %q:%q", ro.Namespace, ro.Name)
		rows = append(rows, parseRules(ro.Namespace, "RO:"+ro.Name, ro.Rules)...)
	}

	return rows, nil
}

func fetchClusterRoleBindings(f Factory) ([]rbacv1.ClusterRoleBinding, error) {
	oo, err := f.List("rbac.authorization.k8s.io/v1/clusterrolebindings", render.ClusterScope, labels.Everything())
	if err != nil {
		return nil, err
	}

	crbs := make([]rbacv1.ClusterRoleBinding, len(oo))
	for i, o := range oo {
		var crb rbacv1.ClusterRoleBinding
		if e := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &crb); e != nil {
			return nil, e
		}
		crbs[i] = crb
	}

	return crbs, nil
}

func fetchRoleBindings(f Factory) ([]rbacv1.RoleBinding, error) {
	oo, err := f.List("rbac.authorization.k8s.io/v1/rolebindings", render.ClusterScope, labels.Everything())
	if err != nil {
		return nil, err
	}

	rbs := make([]rbacv1.RoleBinding, 0, len(oo))
	for _, o := range oo {
		var rb rbacv1.RoleBinding
		if e := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &rb); e != nil {
			return nil, e
		}
		rbs = append(rbs, rb)
	}

	return rbs, nil
}

func (p *Policy) fetchRoleBindingSubjects(kind, name string) ([]string, error) {
	rbs, err := fetchRoleBindings(p.factory)
	if err != nil {
		return nil, err
	}
	ss := make([]string, 0, len(rbs))
	for _, rb := range rbs {
		for _, s := range rb.Subjects {
			if s.Kind == kind && s.Name == name {
				ss = append(ss, rb.RoleRef.Kind+":"+rb.Name)
			}
		}
	}

	return ss, nil
}

func (p *Policy) fetchClusterRoles() ([]rbacv1.ClusterRole, error) {
	oo, err := p.factory.List("rbac.authorization.k8s.io/v1/clusterroles", render.ClusterScope, labels.Everything())
	if err != nil {
		return nil, err
	}

	crs := make([]rbacv1.ClusterRole, len(oo))
	for i, o := range oo {
		var cr rbacv1.ClusterRole
		if e := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cr); e != nil {
			return nil, err
		}
		crs[i] = cr
	}

	return crs, nil
}

func (p *Policy) fetchRoles() ([]rbacv1.Role, error) {
	oo, err := p.factory.List("rbac.authorization.k8s.io/v1/roles", render.AllNamespaces, labels.Everything())
	if err != nil {
		return nil, err
	}

	rr := make([]rbacv1.Role, len(oo))
	for i, o := range oo {
		var ro rbacv1.Role
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ro); err != nil {
			return nil, err
		}
		rr[i] = ro
	}

	return rr, nil
}

func in(nn []string, match string) bool {
	for _, n := range nn {
		if n == match {
			return true
		}
	}
	return false
}

func parseRules(ns, binding string, rules []rbacv1.PolicyRule) render.Policies {
	pp := make(render.Policies, 0, len(rules))
	for _, rule := range rules {
		for _, grp := range rule.APIGroups {
			for _, res := range rule.Resources {
				for _, na := range rule.ResourceNames {
					pp = pp.Upsert(render.NewPolicyRes(ns, binding, FQN(res, na), grp, rule.Verbs))
				}
				pp = pp.Upsert(render.NewPolicyRes(ns, binding, FQN(grp, res), grp, rule.Verbs))
			}
		}
		for _, nres := range rule.NonResourceURLs {
			if nres[0] != '/' {
				nres = "/" + nres
			}
			pp = pp.Upsert(render.NewPolicyRes(ns, binding, nres, "n/a", rule.Verbs))
		}
	}

	return pp
}
