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
	lsel := labels.Everything()
	if sel, e := labels.ConvertSelectorToLabelsMap(strLabel); ok && e == nil {
		lsel = sel.AsSelector()
	}

	const gvr = "apiextensions.k8s.io/v1beta1/customresourcedefinitions"
	oo, err := c.Factory.List(gvr, "-", true, lsel)
	if err != nil {
		return nil, err
	}

	return oo, nil
}
