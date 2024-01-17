// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"

	"github.com/derailed/k9s/internal"
	v1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*CustomResourceDefinition)(nil)
	_ Nuker    = (*CustomResourceDefinition)(nil)
)

// CustomResourceDefinition represents a CRD resource model.
type CustomResourceDefinition struct {
	Resource
}

// IsHappy check for happy deployments.
func (c *CustomResourceDefinition) IsHappy(crd v1.CustomResourceDefinition) bool {
	versions := make([]string, 0, 3)
	for _, v := range crd.Spec.Versions {
		if v.Served && !v.Deprecated {
			versions = append(versions, v.Name)
			break
		}
	}

	return len(versions) > 0
}

// List returns a collection of nodes.
func (c *CustomResourceDefinition) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	strLabel, ok := ctx.Value(internal.KeyLabels).(string)
	labelSel := labels.Everything()
	if sel, e := labels.ConvertSelectorToLabelsMap(strLabel); ok && e == nil {
		labelSel = sel.AsSelector()
	}

	const gvr = "apiextensions.k8s.io/v1/customresourcedefinitions"
	return c.getFactory().List(gvr, "-", false, labelSel)
}
