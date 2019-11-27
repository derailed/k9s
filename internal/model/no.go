package model

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/watch"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var _ render.NodeWithMetrics = &NodeWithMetrics{}

// Node represents a node model.
type Node struct {
	*Resource
}

// NewNode returns a new node model.
func NewNode() *Node {
	return &Node{Resource: NewResource()}
}

// List returns a collection of node resources.
func (n *Node) List(_ string) ([]runtime.Object, error) {
	nn, err := n.factory.Client().DialOrDie().CoreV1().Nodes().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, len(nn.Items))
	for i, no := range nn.Items {
		o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&no)
		if err != nil {
			return nil, err
		}
		oo[i] = &unstructured.Unstructured{Object: o}
	}
	return oo, nil
}

// Hydrate returns nodes as rows.
func (n *Node) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	mx := k8s.NewMetricsServer(n.factory.Client().(k8s.Connection))
	mmx, err := mx.FetchNodesMetrics()
	if err != nil {
		return err
	}

	var index int
	size := len(re.Header(""))
	for _, no := range oo {
		o := no.(*unstructured.Unstructured)
		pods, err := n.nodePods(n.factory, o.Object["metadata"].(map[string]interface{})["name"].(string))
		if err != nil {
			panic(err)
		}
		row := render.Row{Fields: make([]string, size)}
		nmx := NodeWithMetrics{
			o,
			nodeMetricsFor(o, mmx),
			pods,
		}
		if err := re.Render(&nmx, "", &row); err != nil {
			return err
		}
		rr[index] = row
		index++
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

func (n *Node) nodePods(f *watch.Factory, node string) ([]*v1.Pod, error) {
	pp, err := f.List("", "v1/pods", labels.Everything())
	if err != nil {
		return nil, err
	}

	pods := make([]*v1.Pod, 0, len(pp))
	for _, p := range pp {
		o := p.(*unstructured.Unstructured)

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
