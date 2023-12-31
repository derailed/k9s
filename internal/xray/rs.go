// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package xray

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// ReplicaSet represents an xray renderer.
type ReplicaSet struct{}

// Render renders an xray node.
func (r *ReplicaSet) Render(ctx context.Context, ns string, o interface{}) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	var rs appsv1.ReplicaSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &rs)
	if err != nil {
		return err
	}

	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("Expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}

	root := NewTreeNode("apps/v1/replicasets", client.FQN(rs.Namespace, rs.Name))
	oo, err := locatePods(ctx, rs.Namespace, rs.Spec.Selector)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, KeyParent, root)
	var re Pod
	for _, o := range oo {
		p, ok := o.(*unstructured.Unstructured)
		if !ok {
			return fmt.Errorf("expecting *Unstructured but got %T", o)
		}
		if err := re.Render(ctx, ns, &render.PodWithMetrics{Raw: p}); err != nil {
			return err
		}
	}

	if root.IsLeaf() {
		return nil
	}

	gvr, nsID := "v1/namespaces", client.FQN(client.ClusterScope, rs.Namespace)
	nsn := parent.Find(gvr, nsID)
	if nsn == nil {
		nsn = NewTreeNode(gvr, nsID)
		parent.Add(nsn)
	}
	nsn.Add(root)

	return r.validate(root, rs)
}

func (*ReplicaSet) validate(root *TreeNode, rs appsv1.ReplicaSet) error {
	root.Extras[StatusKey] = OkStatus
	var r int32
	if rs.Spec.Replicas != nil {
		r = int32(*rs.Spec.Replicas)
	}
	a := rs.Status.Replicas
	if a != r {
		root.Extras[StatusKey] = ToastStatus
	}
	root.Extras[InfoKey] = fmt.Sprintf("%d/%d", a, r)

	return nil
}
