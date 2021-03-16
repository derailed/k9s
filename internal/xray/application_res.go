package xray

import (
	"context"
	"fmt"

	v1alpha1 "github.com/argoproj/argo-cd/pkg/apis/application/v1alpha1"
	"github.com/derailed/k9s/internal/client"
)

// ApplicationResource represents an xray renderer.
type ApplicationResource struct{}

// Render renders an xray node.
func (a *ApplicationResource) Render(ctx context.Context, ns string, res v1alpha1.ResourceStatus) error {
	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("Expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}

	gvr := gvkToGvr(res.GroupVersionKind())

	root := NewTreeNode(gvr.String(), client.FQN(res.Namespace, res.Name))
	ctx = context.WithValue(ctx, KeyParent, root)

	if res.Namespace == "" {
		parent.Add(root)
	} else {
		gvr, nsID := "v1/namespaces", client.FQN(client.ClusterScope, res.Namespace)
		nsn := parent.Find(gvr, nsID)
		if nsn == nil {
			nsn = NewTreeNode(gvr, nsID)
			parent.Add(nsn)
		}
		nsn.Add(root)
	}

	return nil
}
