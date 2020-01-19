package xray

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

type ServiceAccount struct{}

func (s *ServiceAccount) Render(ctx context.Context, ns string, o interface{}) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("ServiceAccount render expecting *Unstructured, but got %T", o)
	}

	var sa v1.ServiceAccount
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &sa)
	if err != nil {
		return err
	}

	f, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return fmt.Errorf("no factory found in context")
	}

	node := NewTreeNode("v1/serviceaccounts", client.FQN(sa.Namespace, sa.Name))
	node.Extras[StatusKey] = OkStatus
	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("Expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}
	parent.Add(node)

	for _, sec := range sa.Secrets {
		addRef(f, node, "v1/secrets", client.FQN(sa.Namespace, sec.Name), nil)
	}
	for _, sec := range sa.ImagePullSecrets {
		addRef(f, node, "v1/secrets", client.FQN(sa.Namespace, sec.Name), nil)
	}

	return nil
}
