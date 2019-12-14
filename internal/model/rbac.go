package model

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

type Rbac struct {
	Resource
}

func (r *Rbac) List(ctx context.Context) ([]runtime.Object, error) {
	gvr, ok := ctx.Value(internal.KeyGVR).(string)
	if !ok {
		return nil, fmt.Errorf("expecting a context gvr")
	}
	r.gvr = gvr
	path, ok := ctx.Value(internal.KeyPath).(string)
	log.Debug().Msgf("LISTING RBACK %q--%q", r.gvr, path)
	if !ok || path == "" {
		return r.Resource.List(ctx)
	}

	switch client.GVR(r.gvr).ToR() {
	case "clusterrolebindings":
		return r.loadClusterRoleBinding(path)
	case "rolebindings":
		return r.loadRoleBinding(path)
	case "clusterroles":
		return r.loadClusterRole(path)
	case "roles":
		return r.loadRole(path)
	default:
		return nil, fmt.Errorf("expecting clusterrole/role but found %s", client.GVR(r.gvr).ToR())
	}
}

func (r *Rbac) loadClusterRoleBinding(path string) ([]runtime.Object, error) {
	o, err := r.factory.Get("rbac.authorization.k8s.io/v1/clusterrolebindings", path, labels.Everything())
	if err != nil {
		return nil, err
	}

	var crb rbacv1.ClusterRoleBinding
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &crb)
	if err != nil {
		return nil, err
	}

	kind := "rbac.authorization.k8s.io/v1/clusterroles"
	crbo, err := r.factory.Get(kind, client.FQN("-", crb.RoleRef.Name), labels.Everything())
	if err != nil {
		return nil, err
	}
	var cr rbacv1.ClusterRole
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(crbo.(*unstructured.Unstructured).Object, &cr)
	if err != nil {
		return nil, err
	}
	return r.parseRules(cr.Rules), nil
}

func (r *Rbac) loadRoleBinding(path string) ([]runtime.Object, error) {
	o, err := r.factory.Get("rbac.authorization.k8s.io/v1/rolebindings", path, labels.Everything())
	if err != nil {
		return nil, err
	}

	var rb rbacv1.RoleBinding
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &rb)
	if err != nil {
		return nil, err
	}

	if rb.RoleRef.Kind == "ClusterRole" {
		kind := "rbac.authorization.k8s.io/v1/clusterroles"
		o, err := r.factory.Get(kind, client.FQN("-", rb.RoleRef.Name), labels.Everything())
		if err != nil {
			return nil, err
		}
		var cr rbacv1.ClusterRole
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cr)
		if err != nil {
			return nil, err
		}
		return r.parseRules(cr.Rules), nil
	}

	kind := "rbac.authorization.k8s.io/v1/roles"
	ro, err := r.factory.Get(kind, client.FQN(rb.Namespace, rb.RoleRef.Name), labels.Everything())
	if err != nil {
		return nil, err
	}
	var role rbacv1.Role
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(ro.(*unstructured.Unstructured).Object, &role)
	if err != nil {
		return nil, err
	}

	return r.parseRules(role.Rules), nil
}

func (r *Rbac) loadClusterRole(path string) ([]runtime.Object, error) {
	o, err := r.factory.Get("rbac.authorization.k8s.io/v1/clusterroles", path, labels.Everything())
	if err != nil {
		return nil, err
	}

	var cr rbacv1.ClusterRole
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cr)
	if err != nil {
		return nil, err
	}

	return r.parseRules(cr.Rules), nil
}

func (r *Rbac) loadRole(path string) ([]runtime.Object, error) {
	o, err := r.factory.Get("rbac.authorization.k8s.io/v1/roles", path, labels.Everything())
	if err != nil {
		return nil, err
	}

	var ro rbacv1.Role
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ro)
	if err != nil {
		return nil, err
	}

	return r.parseRules(ro.Rules), nil
}

func makeRes(res, grp string, vv []string) *render.PolicyRes {
	return &render.PolicyRes{
		Resource: res,
		Group:    grp,
		Verbs:    vv,
	}
}

func (r *Rbac) parseRules(rules []rbacv1.PolicyRule) []runtime.Object {
	m := make([]runtime.Object, 0, len(rules))
	for _, rule := range rules {
		for _, grp := range rule.APIGroups {
			for _, res := range rule.Resources {
				k := res
				if grp != "" {
					k = res + "." + grp
				}
				for _, na := range rule.ResourceNames {
					m = upsert(m, makeRes(FQN(k, na), grp, rule.Verbs))
				}
				m = upsert(m, makeRes(k, grp, rule.Verbs))
			}
		}
		for _, nres := range rule.NonResourceURLs {
			if nres[0] != '/' {
				nres = "/" + nres
			}
			m = upsert(m, makeRes(nres, "", rule.Verbs))
		}
	}

	return m
}

func upsert(rr []runtime.Object, p *render.PolicyRes) []runtime.Object {
	idx, ok := find(rr, p.Resource)
	if !ok {
		return append(rr, p)
	}
	rr[idx] = p

	return rr
}

// Find locates a row by id. Retturns false is not found.
func find(rr []runtime.Object, res string) (int, bool) {
	for i, r := range rr {
		p := r.(*render.PolicyRes)
		if p.Resource == res {
			return i, true
		}
	}

	return 0, false
}
