package dao

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var (
	_ Accessor = (*Node)(nil)
)

// NodeMetricsFunc retrieves node metrics.
type NodeMetricsFunc func() (*mv1beta1.NodeMetricsList, error)

// Node represents a node model.
type Node struct {
	Resource
}

// List returns a collection of node resources.
func (n *Node) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	labels, ok := ctx.Value(internal.KeyLabels).(string)
	if !ok {
		log.Warn().Msgf("No label selector found in context")
	}

	mx := client.NewMetricsServer(n.Client())
	nmx, err := mx.FetchNodesMetrics()
	if err != nil {
		log.Warn().Err(err).Msgf("No node metrics")
	}

	nn, err := FetchNodes(n.Factory, labels)
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

// ----------------------------------------------------------------------------
// Helpers...

// FetchNodes retrieves all nodes.
func FetchNodes(f Factory, labelsSel string) (*v1.NodeList, error) {
	var list v1.NodeList
	auth, err := f.Client().CanI("", "v1/nodes", []string{client.ListVerb})
	if err != nil {
		return &list, err
	}
	if !auth {
		return &list, fmt.Errorf("user is not authorized to list nodes")
	}

	return f.Client().DialOrDie().CoreV1().Nodes().List(metav1.ListOptions{
		LabelSelector: labelsSel,
	})
}

func nodeMetricsFor(fqn string, mmx *mv1beta1.NodeMetricsList) *mv1beta1.NodeMetrics {
	if mmx == nil {
		return nil
	}
	for _, mx := range mmx.Items {
		if MetaFQN(mx.ObjectMeta) == fqn {
			return &mx
		}
	}
	return nil
}
