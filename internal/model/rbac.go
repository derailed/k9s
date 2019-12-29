package model

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	crbGVR = "rbac.authorization.k8s.io/v1/clusterrolebindings"
	crGVR  = "rbac.authorization.k8s.io/v1/clusterroles"
	rbGVR  = "rbac.authorization.k8s.io/v1/rolebindings"
	rGVR   = "rbac.authorization.k8s.io/v1/roles"
)

// Rbac represents a model for listing rbac resources.
type Rbac struct {
	Resource
}

// List lists out rbac resources.
func (r *Rbac) List(ctx context.Context) ([]runtime.Object, error) {
	gvr, ok := ctx.Value(internal.KeyGVR).(string)
	if !ok {
		return nil, fmt.Errorf("expecting a context gvr")
	}
	r.gvr = gvr
	path, ok := ctx.Value(internal.KeyPath).(string)
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
	o, err := r.factory.Get(crbGVR, path, labels.Everything())
	if err != nil {
		return nil, err
	}

	var crb rbacv1.ClusterRoleBinding
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &crb)
	if err != nil {
		return nil, err
	}

	crbo, err := r.factory.Get(crGVR, client.FQN("-", crb.RoleRef.Name), labels.Everything())
	if err != nil {
		return nil, err
	}
	var cr rbacv1.ClusterRole
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(crbo.(*unstructured.Unstructured).Object, &cr)
	if err != nil {
		return nil, err
	}

	return asRuntimeObjects(parseRules(render.ClusterScope, "-", cr.Rules)), nil
}

func (r *Rbac) loadRoleBinding(path string) ([]runtime.Object, error) {
	o, err := r.factory.Get(rbGVR, path, labels.Everything())
	if err != nil {
		return nil, err
	}

	var rb rbacv1.RoleBinding
	if err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &rb); err != nil {
		return nil, err
	}

	if rb.RoleRef.Kind == "ClusterRole" {
		o, e := r.factory.Get(crGVR, client.FQN("-", rb.RoleRef.Name), labels.Everything())
		if e != nil {
			return nil, e
		}
		var cr rbacv1.ClusterRole
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cr)
		if err != nil {
			return nil, err
		}
		return asRuntimeObjects(parseRules(render.ClusterScope, "-", cr.Rules)), nil
	}

	ro, err := r.factory.Get(rGVR, client.FQN(rb.Namespace, rb.RoleRef.Name), labels.Everything())
	if err != nil {
		return nil, err
	}
	var role rbacv1.Role
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(ro.(*unstructured.Unstructured).Object, &role)
	if err != nil {
		return nil, err
	}

	return asRuntimeObjects(parseRules(render.ClusterScope, "-", role.Rules)), nil
}

func (r *Rbac) loadClusterRole(path string) ([]runtime.Object, error) {
	o, err := r.factory.Get(crGVR, path, labels.Everything())
	if err != nil {
		return nil, err
	}

	var cr rbacv1.ClusterRole
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cr)
	if err != nil {
		return nil, err
	}

	return asRuntimeObjects(parseRules(render.ClusterScope, "-", cr.Rules)), nil
}

func (r *Rbac) loadRole(path string) ([]runtime.Object, error) {
	o, err := r.factory.Get(rGVR, path, labels.Everything())
	if err != nil {
		return nil, err
	}

	var ro rbacv1.Role
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &ro)
	if err != nil {
		return nil, err
	}

	return asRuntimeObjects(parseRules(render.ClusterScope, "-", ro.Rules)), nil
}

func asRuntimeObjects(rr render.Policies) []runtime.Object {
	oo := make([]runtime.Object, len(rr))
	for i, r := range rr {
		oo[i] = r
	}

	return oo
}
