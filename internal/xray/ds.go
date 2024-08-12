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

// DaemonSet represents an xray renderer.
type DaemonSet struct{}

// Render renders an xray node.
func (d *DaemonSet) Render(ctx context.Context, ns string, o interface{}) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	var ds appsv1.DaemonSet
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ds)
	if err != nil {
		return err
	}

	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("Expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}

	root := NewTreeNode("apps/v1/daemonsets", client.FQN(ds.Namespace, ds.Name))
	oo, err := locatePods(ctx, ds.Namespace, ds.Spec.Selector)
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
	gvr, nsID := "v1/namespaces", client.FQN(client.ClusterScope, ds.Namespace)
	nsn := parent.Find(gvr, nsID)
	if nsn == nil {
		nsn = NewTreeNode(gvr, nsID)
		parent.Add(nsn)
	}
	nsn.Add(root)

	return d.validate(root, ds)
}

func (*DaemonSet) validate(root *TreeNode, ds appsv1.DaemonSet) error {
	root.Extras[StatusKey] = OkStatus
	d := ds.Status.DesiredNumberScheduled
	a := ds.Status.NumberAvailable
	if d != a {
		root.Extras[StatusKey] = ToastStatus
	}
	root.Extras[InfoKey] = fmt.Sprintf("%d/%d", a, d)

	return nil
}
