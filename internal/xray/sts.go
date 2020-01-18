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

type StatefulSet struct{}

func (p *StatefulSet) Render(ctx context.Context, ns string, o interface{}) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Unstructured, but got %T", o)
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

	nsID, gvr := client.FQN(client.ClusterScope, sts.Namespace), "v1/namespaces"
	nsn := parent.Find(gvr, nsID)
	if nsn == nil {
		nsn = NewTreeNode(gvr, nsID)
		parent.Add(nsn)
	}
	root := NewTreeNode("apps/v1/deployments", client.FQN(sts.Namespace, sts.Name))
	nsn.Add(root)

	l, err := metav1.LabelSelectorAsSelector(sts.Spec.Selector)
	if err != nil {
		return err
	}

	f, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return fmt.Errorf("Expecting a factory but got %T", ctx.Value(internal.KeyFactory))
	}

	fsel, err := labels.ConvertSelectorToLabelsMap(l.String())
	if err != nil {
		return err
	}

	oo, err := f.List("v1/pods", sts.Namespace, false, fsel.AsSelector())
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

	root.Extras[StatusKey] = OkStatus
	var r int32
	if sts.Spec.Replicas != nil {
		r = int32(*sts.Spec.Replicas)
	}
	a := sts.Status.Replicas
	if a != r {
		root.Extras[StatusKey] = ToastStatus
	}

	return nil
}
