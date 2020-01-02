package model

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// Pod represents a pod model.
type Pod struct {
	Resource
}

// List returns a collection of nodes.
func (p *Pod) List(ctx context.Context) ([]runtime.Object, error) {
	oo, err := p.Resource.List(ctx)
	if err != nil {
		return oo, err
	}

	pmx, ok := ctx.Value(internal.KeyMetrics).(*mv1beta1.PodMetricsList)
	if !ok {
		log.Warn().Msgf("expecting context PodMetricsList")
	}

	sel, ok := ctx.Value(internal.KeyFields).(string)
	if !ok {
		return oo, nil
	}
	fsel, err := labels.ConvertSelectorToLabelsMap(sel)
	if err != nil {
		return nil, err
	}
	nodeName := fsel["spec.nodeName"]

	var res []runtime.Object
	for _, o := range oo {
		u, ok := o.(*unstructured.Unstructured)
		if !ok {
			return res, fmt.Errorf("expecting *unstructured.Unstructured but got `%T", o)
		}
		if nodeName == "" {
			res = append(res, &render.PodWithMetrics{Raw: u, MX: podMetricsFor(o, pmx)})
			continue
		}

		spec, ok := u.Object["spec"].(map[string]interface{})
		if !ok {
			return res, fmt.Errorf("expecting interface map but got `%T", o)
		}
		if spec["nodeName"] == nodeName {
			res = append(res, &render.PodWithMetrics{Raw: u, MX: podMetricsFor(o, pmx)})
		}
	}

	return res, nil
}

// Hydrate returns pod resources as rows.
func (p *Pod) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	var index int
	for _, o := range oo {
		po, ok := o.(*render.PodWithMetrics)
		if !ok {
			return fmt.Errorf("expecting *PodWithMetric but got %T", po)
		}
		var row render.Row
		if err := re.Render(po, p.namespace, &row); err != nil {
			return err
		}
		rr[index] = row
		index++
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func podMetricsFor(o runtime.Object, mmx *mv1beta1.PodMetricsList) *mv1beta1.PodMetrics {
	fqn := extractFQN(o)
	for _, mx := range mmx.Items {
		if MetaFQN(mx.ObjectMeta) == fqn {
			return &mx
		}
	}
	return nil
}
