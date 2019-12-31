package model

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
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
			return res, fmt.Errorf("expecting unstructured but got `%T", o)
		}
		spec, ok := u.Object["spec"].(map[string]interface{})
		if !ok {
			return res, fmt.Errorf("expecting interface map but got `%T", o)
		}
		if nodeName == "" || spec["nodeName"] == nodeName {
			res = append(res, o)
		}
	}

	return res, nil
}

// Hydrate returns pod resources as rows.
func (p *Pod) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	mx := client.NewMetricsServer(p.factory.Client())
	mmx, err := mx.FetchPodsMetrics(p.namespace)
	if err != nil {
		log.Warn().Err(err).Msgf("No metrics found for pod")
	}

	var index int
	for _, o := range oo {
		var (
			row render.Row
			pmx = PodWithMetrics{object: o, mx: podMetricsFor(o, mmx)}
		)
		if err := re.Render(&pmx, p.namespace, &row); err != nil {
			return err
		}
		rr[index] = row
		index++
	}

	return nil
}

func podMetricsFor(o runtime.Object, mmx *mv1beta1.PodMetricsList) *mv1beta1.PodMetrics {
	fqn := extractFQN(o)
	for _, mx := range mmx.Items {
		if MetaFQN(mx.ObjectMeta) == fqn {
			return &mx
		}
	}
	return nil
}

// PodWithMetrics represents a pod and its metrics.
type PodWithMetrics struct {
	object runtime.Object
	mx     *mv1beta1.PodMetrics
}

// Object returns a pod.
func (p *PodWithMetrics) Object() runtime.Object {
	return p.object
}

// Metrics returns the metrics associated with the pod.
func (p *PodWithMetrics) Metrics() *mv1beta1.PodMetrics {
	return p.mx
}
