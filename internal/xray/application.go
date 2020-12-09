package xray

import (
	"context"
	"fmt"
	"strings"

	v1alpha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Application represents an xray renderer.
type Application struct{}

// Render renders an xray node.
func (a *Application) Render(ctx context.Context, ns string, o interface{}) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Unstructured, but got %T", o)
	}

	var app v1alpha1.Application
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &app)
	if err != nil {
		return err
	}

	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("Expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}

	root := NewTreeNode("argoproj.io/v1alpha1/applications", client.FQN(app.Namespace, app.Name))
	ctx = context.WithValue(ctx, KeyParent, root)

	var ar ApplicationResource
	var dp Deployment
	var svc Service
	f, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return fmt.Errorf("Expecting a factory but got %T", ctx.Value(internal.KeyFactory))
	}
	for _, res := range app.Status.Resources {
		gvr := gvkToGvr(res.GroupVersionKind())
		switch gvr.String() {
		case "apps/v1/deployments":
			d, err := f.Get("apps/v1/deployments", fmt.Sprintf("%s/%s", res.Namespace, res.Name), false, labels.Everything())
			if err != nil {
				return err
			}

			if err := dp.Render(ctx, app.Namespace, d); err != nil {
				return err
			}

		case "v1/services":
			d, err := f.Get("v1/services", fmt.Sprintf("%s/%s", res.Namespace, res.Name), false, labels.Everything())
			if err != nil {
				return err
			}

			if err := svc.Render(ctx, app.Namespace, d); err != nil {
				return err
			}

		default:
			if err := ar.Render(ctx, app.Namespace, res); err != nil {
				return err
			}
		}
		/*
			if meta, ok := model.Registry[gvr.String()]; ok {
				if meta.TreeRenderer != nil {
					if err := meta.TreeRenderer.Render(ctx, app.Namespace, res); err != nil {
						return err
					}
				}
			}
		*/
	}

	gvr, nsID := "v1/namespaces", client.FQN(client.ClusterScope, app.Namespace)
	nsn := parent.Find(gvr, nsID)
	if nsn == nil {
		nsn = NewTreeNode(gvr, nsID)
		parent.Add(nsn)
	}
	nsn.Add(root)

	return nil
}

func gvkToGvr(gvk schema.GroupVersionKind) client.GVR {
	gvr := fmt.Sprintf("%s/%ss", gvk.Version, strings.ToLower(gvk.Kind))
	if gvk.Group != "" {
		gvr = fmt.Sprintf("%s/%s", gvk.Group, gvr)
	}
	return client.NewGVR(gvr)
}
