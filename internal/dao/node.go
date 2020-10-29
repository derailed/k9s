package dao

import (
	"context"
	"fmt"
	"io"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	"k8s.io/kubectl/pkg/drain"
	"k8s.io/kubectl/pkg/scheme"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var (
	_ Accessor       = (*Node)(nil)
	_ NodeMaintainer = (*Node)(nil)
)

// NodeMetricsFunc retrieves node metrics.
type NodeMetricsFunc func() (*mv1beta1.NodeMetricsList, error)

// Node represents a node model.
type Node struct {
	Resource
}

// ToggleCordon toggles cordon/uncordon a node.
func (n *Node) ToggleCordon(path string, cordon bool) error {
	log.Debug().Msgf("CORDON %q::%t -- %q", path, cordon, n.gvr.GVK())
	o, err := FetchNode(context.Background(), n.Factory, path)
	if err != nil {
		return err
	}

	h, err := drain.NewCordonHelperFromRuntimeObject(o, scheme.Scheme, n.gvr.GVK())
	if err != nil {
		log.Debug().Msgf("BOOM %v", err)
		return err
	}

	if !h.UpdateIfRequired(cordon) {
		if cordon {
			return fmt.Errorf("node is already cordoned")
		}
		return fmt.Errorf("node is already uncordoned")
	}
	dial, err := n.Factory.Client().Dial()
	if err != nil {
		return err
	}

	err, patchErr := h.PatchOrReplace(dial, false)
	if patchErr != nil {
		return patchErr
	}
	if err != nil {
		return err
	}

	return nil
}

func (o DrainOptions) toDrainHelper(k kubernetes.Interface, w io.Writer) drain.Helper {
	return drain.Helper{
		Client:              k,
		GracePeriodSeconds:  o.GracePeriodSeconds,
		Timeout:             o.Timeout,
		DeleteLocalData:     o.DeleteLocalData,
		IgnoreAllDaemonSets: o.IgnoreAllDaemonSets,
		Out:                 w,
		ErrOut:              w,
	}
}

// Drain drains a node.
func (n *Node) Drain(path string, opts DrainOptions, w io.Writer) error {
	_ = n.ToggleCordon(path, true)

	dial, err := n.Factory.Client().Dial()
	if err != nil {
		return err
	}
	h := opts.toDrainHelper(dial, w)
	dd, errs := h.GetPodsForDeletion(path)
	if len(errs) != 0 {
		for _, e := range errs {
			if _, err := h.ErrOut.Write([]byte(e.Error() + "\n")); err != nil {
				return err
			}
		}
		return errs[0]
	}

	if err := h.DeleteOrEvictPods(dd.Pods()); err != nil {
		return err
	}
	fmt.Fprintf(h.Out, "Node %s drained!", path)

	return nil
}

// Get returns a node resource.
func (n *Node) Get(ctx context.Context, path string) (runtime.Object, error) {
	var (
		nmx *mv1beta1.NodeMetricsList
		err error
	)
	if withMx, ok := ctx.Value(internal.KeyWithMetrics).(bool); withMx || !ok {
		if nmx, err = client.DialMetrics(n.Client()).FetchNodesMetrics(ctx); err != nil {
			log.Debug().Err(err).Msgf("No node metrics")
		}
	}

	no, err := FetchNode(ctx, n.Factory, path)
	if err != nil {
		return nil, err
	}
	o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&no)
	if err != nil {
		return nil, err
	}

	return &render.NodeWithMetrics{
		Raw: &unstructured.Unstructured{Object: o},
		MX:  nodeMetricsFor(MetaFQN(no.ObjectMeta), nmx),
	}, nil
}

// List returns a collection of node resources.
func (n *Node) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	var (
		nmx *mv1beta1.NodeMetricsList
		err error
	)
	if withMx, ok := ctx.Value(internal.KeyWithMetrics).(bool); withMx || !ok {
		if nmx, err = client.DialMetrics(n.Client()).FetchNodesMetrics(ctx); err != nil {
			log.Warn().Err(err).Msgf("No node metrics")
		}
	}

	labels, _ := ctx.Value(internal.KeyLabels).(string)
	nn, err := FetchNodes(ctx, n.Factory, labels)
	if err != nil {
		return nil, err
	}
	oo := make([]runtime.Object, len(nn.Items))
	for i, no := range nn.Items {
		o, err := runtime.DefaultUnstructuredConverter.ToUnstructured(&nn.Items[i])
		if err != nil {
			return nil, err
		}
		meta, ok := o["metadata"].(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("expecting interface map but got `%T", o)
		}
		pods, err := n.GetPods(meta["name"].(string))
		if err != nil {
			return nil, err
		}
		oo[i] = &render.NodeWithMetrics{
			Raw:  &unstructured.Unstructured{Object: o},
			MX:   nodeMetricsFor(MetaFQN(no.ObjectMeta), nmx),
			Pods: pods,
		}
	}

	return oo, nil
}

// CountPods counts the pods scheduled on a given node.
func (n *Node) CountPods(nodeName string) (int, error) {
	var count int
	oo, err := n.Factory.List("v1/pods", client.AllNamespaces, false, labels.Everything())
	if err != nil {
		return 0, err
	}

	for _, o := range oo {
		u, ok := o.(*unstructured.Unstructured)
		if !ok {
			return count, fmt.Errorf("expecting *unstructured.Unstructured but got `%T", o)
		}
		spec, ok := u.Object["spec"].(map[string]interface{})
		if !ok {
			return count, fmt.Errorf("expecting interface map but got `%T", o)
		}
		if spec["nodeName"] == nodeName {
			count++
		}
	}

	return count, nil
}

// GetPods returns all pods running on given node.
func (n *Node) GetPods(nodeName string) ([]*v1.Pod, error) {
	oo, err := n.Factory.List("v1/pods", client.AllNamespaces, false, labels.Everything())
	if err != nil {
		return nil, err
	}

	pp := make([]*v1.Pod, 0, len(oo))
	for _, o := range oo {
		po := new(v1.Pod)
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, po); err != nil {
			return nil, err
		}
		if po.Spec.NodeName == nodeName {
			pp = append(pp, po)
		}
	}

	return pp, nil
}

// ----------------------------------------------------------------------------
// Helpers...

// FetchNode retrieves a node.
func FetchNode(ctx context.Context, f Factory, path string) (*v1.Node, error) {
	auth, err := f.Client().CanI(client.ClusterScope, "v1/nodes", []string{"get"})
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("user is not authorized to list nodes")
	}

	dial, err := f.Client().Dial()
	if err != nil {
		return nil, err
	}
	return dial.CoreV1().Nodes().Get(ctx, path, metav1.GetOptions{})
}

// FetchNodes retrieves all nodes.
func FetchNodes(ctx context.Context, f Factory, labelsSel string) (*v1.NodeList, error) {
	auth, err := f.Client().CanI(client.ClusterScope, "v1/nodes", []string{client.ListVerb})
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("user is not authorized to list nodes")
	}

	dial, err := f.Client().Dial()
	if err != nil {
		return nil, err
	}
	return dial.CoreV1().Nodes().List(ctx, metav1.ListOptions{
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
