package xray

import (
	"context"
	"fmt"

	v1alpha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// AppProject represents an xray renderer.
type AppProject struct{}

// Render renders an xray node.
func (a *AppProject) Render(ctx context.Context, ns string, o interface{}) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Unstructured, but got %T", o)
	}

	var proj v1alpha1.AppProject
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &proj)
	if err != nil {
		return err
	}

	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("Expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}

	root := NewTreeNode("argoproj.io/v1alpha1/appprojects", client.FQN(proj.Namespace, proj.Name))
	ctx = context.WithValue(ctx, KeyParent, root)

	f, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return fmt.Errorf("Expecting a factory but got %T", ctx.Value(internal.KeyFactory))
	}

	oo, err := f.List("argoproj.io/v1alpha1/applications", "", false, labels.Everything())
	if err != nil {
		return err
	}
	for _, o := range oo {
		a, ok := o.(*unstructured.Unstructured)
		if !ok {
			return fmt.Errorf("expecting *Unstructured but got %T", o)
		}
		var aa v1alpha1.Application
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(a.Object, &aa); err != nil {
			return err
		}
		if aa.Spec.Project != proj.Name {
			continue
		}
		var app Application
		if err := app.Render(ctx, proj.Namespace, a); err != nil {
			return err
		}

	}

	gvr, nsID := "v1/namespaces", client.FQN(client.ClusterScope, proj.Namespace)
	nsn := parent.Find(gvr, nsID)
	if nsn == nil {
		nsn = NewTreeNode(gvr, nsID)
		parent.Add(nsn)
	}
	nsn.Add(root)

	return nil
}
