package xray

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type Namespace struct{}

func (p *Namespace) Render(ctx context.Context, ns string, o interface{}) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected NamespaceWithMetrics, but got %T", o)
	}

	var nss v1.Namespace
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &nss)
	if err != nil {
		return err
	}

	root := NewTreeNode("v1/namespaces", client.FQN(client.ClusterScope, nss.Name))
	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("Expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}
	parent.Add(root)

	return nil
}
