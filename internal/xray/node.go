// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s
package xray

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Node represents an xray renderer for Kubernetes nodes.
type Node struct{}

// Render renders an xray node for a Kubernetes Node, showing all related pods (and their refs) scheduled on it.
func (n *Node) Render(ctx context.Context, _ string, o any) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}

	var no v1.Node
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &no); err != nil {
		return err
	}

	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}

	f, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return fmt.Errorf("expecting a factory but got %T", ctx.Value(internal.KeyFactory))
	}

	root := NewTreeNode(client.NodeGVR, client.FQN(client.ClusterScope, no.Name))

	// List pods scheduled on this node across all namespaces using field selector via DAO.
	pdao := new(dao.Pod)
	pdao.Init(f, client.PodGVR)
	ctxPods := context.WithValue(ctx, internal.KeyFields, "spec.nodeName="+no.Name)
	ctxPods = context.WithValue(ctxPods, KeyParent, root)

	oo, err := pdao.List(ctxPods, client.NamespaceAll)
	if err != nil {
		return err
	}

	var pre Pod
	for _, obj := range oo {
		pwm, ok := obj.(*render.PodWithMetrics)
		if !ok {
			return fmt.Errorf("expected PodWithMetrics, but got %T", obj)
		}
		if err := pre.Render(ctxPods, "", pwm); err != nil {
			return err
		}
	}

	parent.Add(root)
	return nil
}


