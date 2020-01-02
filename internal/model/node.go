package model

import (
	"context"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type NodeMetricsFunc func() (*mv1beta1.NodeMetricsList, error)

// Node represents a node model.
type Node struct {
	Resource
}

// List returns a collection of node resources.
func (n *Node) List(ctx context.Context) ([]runtime.Object, error) {
	defer func(t time.Time) {
		log.Debug().Msgf("LIST NODES elapsed %v", time.Since(t))
	}(time.Now())

	nmx, ok := ctx.Value(internal.KeyMetrics).(*mv1beta1.NodeMetricsList)
	if !ok {
		log.Warn().Msgf("No node metrics available in context")
	}

	nn, err := dao.FetchNodes(n.factory)
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, len(nn.Items))
	for i, no := range nn.Items {
		o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&nn.Items[i])
		if err != nil {
			return nil, err
		}
		oo[i] = &render.NodeWithMetrics{
			Raw: &unstructured.Unstructured{Object: o},
			MX:  nodeMetricsFor(MetaFQN(no.ObjectMeta), nmx),
		}
	}

	return oo, nil
}

// Hydrate returns nodes as rows.
func (n *Node) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	defer func(t time.Time) {
		log.Debug().Msgf("HYDRATE NODES elapsed %v", time.Since(t))
	}(time.Now())

	for i, o := range oo {
		nmx, ok := o.(*render.NodeWithMetrics)
		if !ok {
			return fmt.Errorf("expecting *NodeWithMetrics but got %T", o)
		}

		var row render.Row
		if err := re.Render(nmx, render.ClusterScope, &row); err != nil {
			return err
		}
		rr[i] = row
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func nodeMetricsFor(fqn string, mmx *mv1beta1.NodeMetricsList) *mv1beta1.NodeMetrics {
	for _, mx := range mmx.Items {
		if MetaFQN(mx.ObjectMeta) == fqn {
			return &mx
		}
	}
	return nil
}
