package model

import (
	"context"

	"github.com/derailed/k9s/internal"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// HorizontalPodAutoscaler represents a HPA resource model.
type HorizontalPodAutoscaler struct {
	Resource
}

// List returns a collection of nodes.
func (c *HorizontalPodAutoscaler) List(ctx context.Context) ([]runtime.Object, error) {
	strLabel, ok := ctx.Value(internal.KeyLabels).(string)
	lsel := labels.Everything()
	if sel, err := labels.ConvertSelectorToLabelsMap(strLabel); ok && err == nil {
		lsel = sel.AsSelector()
	}

	gvr := "autoscaling/v2beta2/horizontalpodautoscalers"
	ooV2b2, err := c.factory.List(gvr, c.namespace, lsel)
	if err != nil {
		return nil, err
	}
	if len(ooV2b2) > 0 {
		return ooV2b2, nil
	}

	gvr = "autoscaling/v2beta1/horizontalpodautoscalers"
	ooV2b1, err := c.factory.List(gvr, c.namespace, lsel)
	if err != nil {
		return nil, err
	}
	if len(ooV2b1) > 0 {
		return ooV2b1, nil
	}

	gvr = "autoscaling/v1/horizontalpodautoscalers"
	oo, err := c.factory.List(gvr, c.namespace, lsel)
	if err != nil {
		return nil, err
	}
	return oo, nil
}
