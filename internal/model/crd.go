package model

import (
	"context"

	"github.com/derailed/k9s/internal"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// CustomResourceDefinition represents a CRD resource model.
type CustomResourceDefinition struct {
	Resource
}

// List returns a collection of nodes.
func (c *CustomResourceDefinition) List(ctx context.Context) ([]runtime.Object, error) {
	strLabel, ok := ctx.Value(internal.KeyLabels).(string)
	lsel := labels.Everything()
	if sel, err := labels.ConvertSelectorToLabelsMap(strLabel); ok && err == nil {
		lsel = sel.AsSelector()
	}

	gvr := "apiextensions.k8s.io/v1beta1/customresourcedefinitions"
	oo, err := c.factory.List(gvr, "-", lsel)
	if err != nil {
		return nil, err
	}

	return oo, nil
}
