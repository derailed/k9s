package dao

import (
	"context"

	"github.com/derailed/k9s/internal"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*HorizontalPodAutoscaler)(nil)
	_ Nuker    = (*HorizontalPodAutoscaler)(nil)
)

// HorizontalPodAutoscaler represents a HPA resource model.
type HorizontalPodAutoscaler struct {
	Resource
}

// List returns a collection of nodes.
func (h *HorizontalPodAutoscaler) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	strLabel, ok := ctx.Value(internal.KeyLabels).(string)
	lsel := labels.Everything()
	if sel, err := labels.ConvertSelectorToLabelsMap(strLabel); ok && err == nil {
		lsel = sel.AsSelector()
	}

	rev, err := h.Factory.Client().ServerVersion()
	if err != nil {
		return nil, err
	}

	gvr := "autoscaling/v1/horizontalpodautoscalers"
	if rev.Minor >= "23" {
		gvr = "autoscaling/v2/horizontalpodautoscalers"
	}

	return h.list(gvr, ns, lsel)
}

func (h *HorizontalPodAutoscaler) list(gvr, ns string, sel labels.Selector) ([]runtime.Object, error) {
	oo, err := h.Factory.List(gvr, ns, true, sel)
	if err != nil {
		return nil, err
	}
	return oo, nil
}
