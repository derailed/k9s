package dao

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/rs/zerolog/log"
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

func (h *HorizontalPodAutoscaler) Create(ctx context.Context, _ runtime.Object) (runtime.Object, error) {
	panic("NYI")
}

// List returns a collection of nodes.
func (h *HorizontalPodAutoscaler) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	strLabel, ok := ctx.Value(internal.KeyLabels).(string)
	lsel := labels.Everything()
	if sel, err := labels.ConvertSelectorToLabelsMap(strLabel); ok && err == nil {
		lsel = sel.AsSelector()
	}

	gvrs := []string{
		"autoscaling/v2beta2/horizontalpodautoscalers",
		"autoscaling/v2beta1/horizontalpodautoscalers",
		"autoscaling/v1/horizontalpodautoscalers",
	}

	for _, gvr := range gvrs {
		oo, err := h.list(gvr, ns, lsel)
		if err == nil && len(oo) > 0 {
			return oo, nil
		}
	}
	log.Error().Err(fmt.Errorf("No results for any known HPA versions"))

	return []runtime.Object{}, nil
}

func (h *HorizontalPodAutoscaler) list(gvr, ns string, sel labels.Selector) ([]runtime.Object, error) {
	oo, err := h.Factory.List(gvr, ns, true, sel)
	if err != nil {
		return nil, err
	}
	return oo, nil
}
