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

// StatefulSet represents an xray renderer.
type StatefulSet struct{}

// Render renders an xray node.
func (s *StatefulSet) Render(ctx context.Context, ns string, o interface{}) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	var sts appsv1.StatefulSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &sts)
	if err != nil {
		return err
	}

	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("Expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}

	root := NewTreeNode("apps/v1/statefulsets", client.FQN(sts.Namespace, sts.Name))
	oo, err := locatePods(ctx, sts.Namespace, sts.Spec.Selector)
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

	gvr, nsID := "v1/namespaces", client.FQN(client.ClusterScope, sts.Namespace)
	nsn := parent.Find(gvr, nsID)
	if nsn == nil {
		nsn = NewTreeNode(gvr, nsID)
		parent.Add(nsn)
	}
	nsn.Add(root)

	return s.validate(root, sts)
}

func (*StatefulSet) validate(root *TreeNode, sts appsv1.StatefulSet) error {
	root.Extras[StatusKey] = OkStatus
	var r int32
	if sts.Spec.Replicas != nil {
		r = int32(*sts.Spec.Replicas)
	}
	a := sts.Status.CurrentReplicas
	if a != r {
		root.Extras[StatusKey] = ToastStatus
	}
	root.Extras[InfoKey] = fmt.Sprintf("%d/%d", a, r)

	return nil
}
