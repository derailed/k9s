// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*Policy)(nil)
	_ Nuker    = (*Policy)(nil)
)

// Policy represent rbac policy.
type Policy struct {
	Resource
}

// List returns available policies.
func (p *Policy) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	kind, ok := ctx.Value(internal.KeySubjectKind).(string)
	if !ok {
		return nil, fmt.Errorf("expecting a context subject kind")
	}
	name, ok := ctx.Value(internal.KeySubjectName).(string)
	if !ok {
		return nil, fmt.Errorf("expecting a context subject name")
	}

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
	crbs, err := fetchClusterRoleBindings(p.Factory)
	if err != nil {
		return nil, err
	}

	ns, n := client.Namespaced(name)
	var nn []string
	for _, crb := range crbs {
		for _, s := range crb.Subjects {
			s := s
			if isSameSubject(kind, ns, n, &s) {
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
		if !inList(nn, cr.Name) {
			continue
		}
		rows = append(rows, parseRules(client.NotNamespaced, "CR:"+cr.Name, cr.Rules)...)
	}

	return rows, nil
}

func (p *Policy) loadRoleBinding(kind, name string) (render.Policies, error) {
	rbsMap, err := p.fetchRoleBindingNamespaces(kind, name)
	if err != nil {
		return nil, err
	}
	crs, err := p.fetchClusterRoles()
	if err != nil {
		return nil, err
	}
	rows := make(render.Policies, 0, len(crs))
	for _, cr := range crs {
		if rbNs, ok := rbsMap["ClusterRole:"+cr.Name]; ok {
			log.Debug().Msgf("Loading rules for clusterrole %q:%q", rbNs, cr.Name)
			rows = append(rows, parseRules(rbNs, "CR:"+cr.Name, cr.Rules)...)
		}
	}

	ros, err := p.fetchRoles()
	if err != nil {
		return nil, err
	}
	for _, ro := range ros {
		if _, ok := rbsMap["Role:"+ro.Name]; !ok {
			continue
		}
		log.Debug().Msgf("Loading rules for role %q:%q", ro.Namespace, ro.Name)
		rows = append(rows, parseRules(ro.Namespace, "RO:"+ro.Name, ro.Rules)...)
	}

	return rows, nil
}

func fetchClusterRoleBindings(f Factory) ([]rbacv1.ClusterRoleBinding, error) {
	oo, err := f.List(crbGVR, client.ClusterScope, false, labels.Everything())
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
	oo, err := f.List(rbGVR, client.ClusterScope, false, labels.Everything())
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

func (p *Policy) fetchRoleBindingNamespaces(kind, name string) (map[string]string, error) {
	rbs, err := fetchRoleBindings(p.Factory)
	if err != nil {
		return nil, err
	}

	ns, n := client.Namespaced(name)
	ss := make(map[string]string, len(rbs))
	for _, rb := range rbs {
		for _, s := range rb.Subjects {
			s := s
			if isSameSubject(kind, ns, n, &s) {
				ss[rb.RoleRef.Kind+":"+rb.RoleRef.Name] = rb.Namespace
			}
		}
	}

	return ss, nil
}

// isSameSubject verifies if the incoming type name and namespace match a subject from a
// cluster/roleBinding. A ServiceAccount will always have a namespace and needs to be validated to ensure
// we don't display permissions for a ServiceAccount with the same name in a different namespace
func isSameSubject(kind, ns, name string, subject *rbacv1.Subject) bool {
	if subject.Kind != kind || subject.Name != name {
		return false
	}
	if kind == rbacv1.ServiceAccountKind {
		// Kind and name were checked above, check the namespace
		return client.IsAllNamespaces(ns) || subject.Namespace == ns
	}
	return true
}

func (p *Policy) fetchClusterRoles() ([]rbacv1.ClusterRole, error) {
	const gvr = "rbac.authorization.k8s.io/v1/clusterroles"

	oo, err := p.getFactory().List(gvr, client.ClusterScope, false, labels.Everything())
	if err != nil {
		return nil, err
	}

	crs := make([]rbacv1.ClusterRole, len(oo))
	for i, o := range oo {
		var cr rbacv1.ClusterRole
		if e := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cr); e != nil {
			return nil, e
		}
		crs[i] = cr
	}

	return crs, nil
}

func (p *Policy) fetchRoles() ([]rbacv1.Role, error) {
	const gvr = "rbac.authorization.k8s.io/v1/roles"

	oo, err := p.getFactory().List(gvr, client.BlankNamespace, false, labels.Everything())
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
