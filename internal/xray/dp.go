package xray

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

type Deployment struct{}

func (d *Deployment) Render(ctx context.Context, ns string, o interface{}) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Unstructured, but got %T", o)
	}
	var dp appsv1.Deployment
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &dp)
	if err != nil {
		return err
	}

	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("Expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}

	nsID, gvr := client.FQN(client.ClusterScope, dp.Namespace), "v1/namespaces"
	nsn := parent.Find(gvr, nsID)
	if nsn == nil {
		nsn = NewTreeNode(gvr, nsID)
		parent.Add(nsn)
	}
	root := NewTreeNode("apps/v1/deployments", client.FQN(dp.Namespace, dp.Name))
	nsn.Add(root)

	oo, err := locatePods(ctx, dp.Namespace, dp.Spec.Selector)
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

	return d.validate(root, dp)
}

func (*Deployment) validate(root *TreeNode, dp appsv1.Deployment) error {
	root.Extras[StatusKey] = OkStatus
	var r int32
	if dp.Spec.Replicas != nil {
		r = int32(*dp.Spec.Replicas)
	}
	a := dp.Status.AvailableReplicas
	if a != r {
		root.Extras[StatusKey] = ToastStatus
	}
	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func locatePods(ctx context.Context, ns string, sel *metav1.LabelSelector) ([]runtime.Object, error) {
	l, err := metav1.LabelSelectorAsSelector(sel)
	if err != nil {
		return nil, err
	}
	fsel, err := labels.ConvertSelectorToLabelsMap(l.String())
	if err != nil {
		return nil, err
	}

	f, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return nil, fmt.Errorf("Expecting a factory but got %T", ctx.Value(internal.KeyFactory))
	}

	return f.List("v1/pods", ns, false, fsel.AsSelector())
}
