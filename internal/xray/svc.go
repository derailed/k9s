// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

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

// Service represents an xray renderer.
type Service struct{}

// Render renders an xray node.
func (s *Service) Render(ctx context.Context, ns string, o interface{}) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
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

	root := NewTreeNode("v1/services", client.FQN(svc.Namespace, svc.Name))
	oo, err := s.locatePods(ctx, svc.Namespace, svc.Spec.Selector)
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
	root.Extras[StatusKey] = OkStatus

	if root.IsLeaf() {
		return nil
	}
	gvr, nsID := "v1/namespaces", client.FQN(client.ClusterScope, svc.Namespace)
	nsn := parent.Find(gvr, nsID)
	if nsn == nil {
		nsn = NewTreeNode(gvr, nsID)
		parent.Add(nsn)
	}
	nsn.Add(root)

	return nil
}

func (s *Service) locatePods(ctx context.Context, ns string, sel map[string]string) ([]runtime.Object, error) {
	f, ok := ctx.Value(internal.KeyFactory).(dao.Factory)
	if !ok {
		return nil, fmt.Errorf("Expecting a factory but got %T", ctx.Value(internal.KeyFactory))
	}

	ll := make([]string, 0, len(sel))
	for k, v := range sel {
		ll = append(ll, fmt.Sprintf("%s=%s", k, v))
	}

	fsel, err := labels.ConvertSelectorToLabelsMap(strings.Join(ll, ","))
	if err != nil {
		return nil, err
	}

	return f.List("v1/pods", ns, false, fsel.AsSelector())
}
