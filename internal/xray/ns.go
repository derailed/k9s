// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package xray

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Namespace represents an xray renderer.
type Namespace struct{}

// Render renders an xray node.
func (n *Namespace) Render(ctx context.Context, ns string, o interface{}) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected NamespaceWithMetrics, but got %T", o)
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

	return n.validate(root, nss)
}

func (*Namespace) validate(root *TreeNode, ns v1.Namespace) error {
	root.Extras[StatusKey] = OkStatus
	if ns.Status.Phase == v1.NamespaceTerminating {
		root.Extras[StatusKey] = ToastStatus
	}

	return nil
}
