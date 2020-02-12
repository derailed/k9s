package dao

import (
	"context"

	"github.com/derailed/k9s/internal"
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

// List returns a collection of nodes.
func (c *CustomResourceDefinition) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	strLabel, ok := ctx.Value(internal.KeyLabels).(string)
	labelSel := labels.Everything()
	if sel, e := labels.ConvertSelectorToLabelsMap(strLabel); ok && e == nil {
		labelSel = sel.AsSelector()
	}

	const gvr = "apiextensions.k8s.io/v1beta1/customresourcedefinitions"
	return c.Factory.List(gvr, "-", true, labelSel)
}
