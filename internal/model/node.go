package model

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var _ render.NodeWithMetrics = &NodeWithMetrics{}

// Node represents a node model.
type Node struct {
	Resource
}

// List returns a collection of node resources.
func (n *Node) List(_ context.Context) ([]runtime.Object, error) {
	nn, err := dao.FetchNodes(n.factory)
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, len(nn.Items))
	for i := range nn.Items {
		o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&nn.Items[i])
		if err != nil {
			return nil, err
		}
		oo[i] = &unstructured.Unstructured{Object: o}
	}

	return oo, nil
}

func nameFromMeta(m map[string]interface{}) string {
	meta, ok := m["metadata"].(map[string]interface{})
	if !ok {
		return "n/a"
	}

	name, ok := meta["name"].(string)
	if !ok {
		return "n/a"
	}

	return name
}

// Hydrate returns nodes as rows.
func (n *Node) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	mx := client.NewMetricsServer(n.factory.Client())
	mmx, err := mx.FetchNodesMetrics()
	if err != nil {
		log.Warn().Err(err).Msg("No node metrics")
	}

	for i, o := range oo {
		no, ok := o.(*unstructured.Unstructured)
		if !ok {
			return fmt.Errorf("expecting unstructured but got %T", o)
		}
		pods, err := n.nodePods(n.factory, nameFromMeta(no.Object))
		if err != nil {
			return err
		}

		var (
			row render.Row
			nmx = NodeWithMetrics{
				object: no,
				mx:     nodeMetricsFor(o, mmx),
				pods:   pods,
			}
		)
		if err := re.Render(&nmx, "", &row); err != nil {
			return err
		}
		rr[i] = row
	}

	return nil
}

func nodeMetricsFor(o runtime.Object, mmx *mv1beta1.NodeMetricsList) *mv1beta1.NodeMetrics {
	fqn := extractFQN(o)
	for _, mx := range mmx.Items {
		if MetaFQN(mx.ObjectMeta) == fqn {
			return &mx
		}
	}
	return nil
}

func (n *Node) nodePods(f dao.Factory, node string) ([]*v1.Pod, error) {
	pp, err := f.List("v1/pods", render.AllNamespaces, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	pods := make([]*v1.Pod, 0, len(pp))
	for _, p := range pp {
		o, ok := p.(*unstructured.Unstructured)
		if !ok {
			return nil, fmt.Errorf("expecting unstructured but got %T", p)
		}
		var pod v1.Pod
		err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.Object, &pod)
		if err != nil {
			log.Error().Err(err).Msg("Converting Pod")
			return nil, err
		}
		if pod.Spec.NodeName != node || pod.Status.Phase != v1.PodSucceeded {
			continue
		}
		pods = append(pods, &pod)
	}

	return pods, nil
}

// ----------------------------------------------------------------------------
// Helpers...

// NodeWithMetrics represents a node with its associated metrics.
type NodeWithMetrics struct {
	object runtime.Object
	mx     *mv1beta1.NodeMetrics
	pods   []*v1.Pod
}

// Object returns a node.
func (n *NodeWithMetrics) Object() runtime.Object {
	return n.object
}

// Metrics returns the node metrics.
func (n *NodeWithMetrics) Metrics() *mv1beta1.NodeMetrics {
	return n.mx
}

// Pods return pods running on this node.
func (n *NodeWithMetrics) Pods() []*v1.Pod {
	return n.pods
}
