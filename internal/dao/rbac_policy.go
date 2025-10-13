// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/slogs"
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
func (p *Policy) List(ctx context.Context, _ string) ([]runtime.Object, error) {
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
	for i := range crbs {
		for _, s := range crbs[i].Subjects {
			if isSameSubject(kind, ns, crbs[i].Namespace, n, &s) {
				nn = append(nn, crbs[i].RoleRef.Name)
			}
		}
	}
	crs, err := p.fetchClusterRoles()
	if err != nil {
		return nil, err
	}

	rows := make(render.Policies, 0, len(nn))
	for i := range crs {
		if !inList(nn, crs[i].Name) {
			continue
		}
		rows = append(rows, parseRules(client.NotNamespaced, "CR:"+crs[i].Name, crs[i].Rules)...)
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
	for i := range crs {
		if rbNs, ok := rbsMap["ClusterRole:"+crs[i].Name]; ok {
			slog.Debug("Loading rules for clusterrole",
				slogs.Namespace, rbNs,
				slogs.ResName, crs[i].Name,
			)
			rows = append(rows, parseRules(rbNs, "CR:"+crs[i].Name, crs[i].Rules)...)
		}
	}

	ros, err := p.fetchRoles()
	if err != nil {
		return nil, err
	}
	for i := range ros {
		if _, ok := rbsMap["Role:"+ros[i].Name]; !ok {
			continue
		}
		slog.Debug("Loading rules for role",
			slogs.Namespace, ros[i].Namespace,
			slogs.ResName, ros[i].Name,
		)
		rows = append(rows, parseRules(ros[i].Namespace, "RO:"+ros[i].Name, ros[i].Rules)...)
	}

	return rows, nil
}

func fetchClusterRoleBindings(f Factory) ([]rbacv1.ClusterRoleBinding, error) {
	oo, err := f.List(client.CrbGVR, client.ClusterScope, false, labels.Everything())
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
	oo, err := f.List(client.RobGVR, client.ClusterScope, false, labels.Everything())
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
	for i := range rbs {
		for _, s := range rbs[i].Subjects {
			if isSameSubject(kind, ns, rbs[i].Namespace, n, &s) {
				ss[rbs[i].RoleRef.Kind+":"+rbs[i].RoleRef.Name] = rbs[i].Namespace
			}
		}
	}

	return ss, nil
}

// isSameSubject verifies if the incoming type name and namespace match a subject from a
// cluster/roleBinding. A ServiceAccount will always have a namespace and needs to be validated to ensure
// we don't display permissions for a ServiceAccount with the same name in a different namespace
func isSameSubject(kind, ns, bns, name string, subject *rbacv1.Subject) bool {
	if subject.Kind != kind || subject.Name != name {
		return false
	}
	if kind == rbacv1.ServiceAccountKind {
		// Kind and name were checked above, check the namespace
		cns := subject.Namespace
		if cns == "" {
			cns = bns
		}
		return client.IsAllNamespaces(ns) || cns == ns
	}
	return true
}

func (p *Policy) fetchClusterRoles() ([]rbacv1.ClusterRole, error) {
	oo, err := p.getFactory().List(client.CrGVR, client.ClusterScope, false, labels.Everything())
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
	oo, err := p.getFactory().List(client.RoGVR, client.BlankNamespace, false, labels.Everything())
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
