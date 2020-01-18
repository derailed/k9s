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

type DaemonSet struct{}

func (d *DaemonSet) Render(ctx context.Context, ns string, o interface{}) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Unstructured, but got %T", o)
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

	nsID, gvr := client.FQN(client.ClusterScope, ds.Namespace), "v1/namespaces"
	nsn := parent.Find(gvr, nsID)
	if nsn == nil {
		nsn = NewTreeNode(gvr, nsID)
		parent.Add(nsn)
	}
	root := NewTreeNode("apps/v1/daemonset", client.FQN(ds.Namespace, ds.Name))
	nsn.Add(root)

	oo, err := locatePods(ctx, ds.Namespace, ds.Spec.Selector)
	if err != nil {
		return err
	}

	ctx = context.WithValue(ctx, KeyParent, root)
	var re Pod
	for _, o := range oo {
		p := o.(*unstructured.Unstructured)
		if err := re.Render(ctx, ns, &render.PodWithMetrics{Raw: p}); err != nil {
			return err
		}
	}

	return d.validate(root, ds)
}

func (*DaemonSet) validate(root *TreeNode, ds appsv1.DaemonSet) error {
	root.Extras[StatusKey] = OkStatus
	d := ds.Status.DesiredNumberScheduled
	a := ds.Status.NumberAvailable
	if d != a {
		root.Extras[StatusKey] = ToastStatus
	}

	return nil
}
