package xray

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

type Service struct{}

func (s *Service) Render(ctx context.Context, ns string, o interface{}) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("Expected Unstructured, but got %T", o)
	}

	var svc v1.Service
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &svc)
	if err != nil {
		return err
	}

	parent, ok := ctx.Value(KeyParent).(*TreeNode)
	if !ok {
		return fmt.Errorf("Expecting a TreeNode but got %T", ctx.Value(KeyParent))
	}

	nsID, gvr := client.FQN(client.ClusterScope, svc.Namespace), "v1/namespaces"
	nsn := parent.Find(gvr, nsID)
	if nsn == nil {
		nsn = NewTreeNode(gvr, nsID)
		parent.Add(nsn)
	}
	root := NewTreeNode("apps/v1/services", client.FQN(svc.Namespace, svc.Name))
	nsn.Add(root)

	oo, err := s.locatePods(ctx, svc.Namespace, svc.Spec.Selector)
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

	return nil
}

func (s *Service) locatePods(ctx context.Context, ns string, sel map[string]string) ([]runtime.Object, error) {
	f, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return nil, fmt.Errorf("Expecting a factory but got %T", ctx.Value(internal.KeyFactory))
	}

	var ll []string
	for k, v := range sel {
		ll = append(ll, fmt.Sprintf("%s=%s", k, v))
	}

	fsel, err := labels.ConvertSelectorToLabelsMap(strings.Join(ll, ","))
	if err != nil {
		return nil, err
	}

	return f.List("v1/pods", ns, false, fsel.AsSelector())
}
