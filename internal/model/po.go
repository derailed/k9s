package model

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/render"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// Pod represents a pod model.
type Pod struct {
	*Resource
}

// NewPod returns a new pod model.
func NewPod() *Pod {
	return &Pod{NewResource()}
}

func (p *Pod) FetchContainers(sel string, includeInit bool) ([]string, error) {
	o, err := p.factory.Get(p.namespace, p.gvr, sel, labels.Everything())
	if err != nil {
		return nil, err
	}

	var po v1.Pod
	if runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &po); err != nil {
		return nil, err
	}

	cc := make([]string, 0, len(po.Spec.Containers))
	for _, c := range po.Spec.Containers {
		cc = append(cc, c.Name)
	}

	if includeInit {
		for _, c := range po.Spec.InitContainers {
			cc = append(cc, c.Name)
		}
	}

	return cc, nil
}

// Render returns pod resources as rows.
func (p *Pod) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	mx := k8s.NewMetricsServer(p.factory.Client().(k8s.Connection))
	mmx, err := mx.FetchPodsMetrics(p.namespace)
	if err != nil {
		return err
	}

	var index int
	size := len(re.Header(p.namespace))
	for _, o := range oo {
		row := render.Row{Fields: make([]string, size)}
		pmx := PodWithMetrics{o, podMetricsFor(o, mmx)}
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
