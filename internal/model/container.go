package model

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
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var _ render.ContainerWithMetrics = &ContainerWithMetrics{}

// Container represents a container model.
type Container struct {
	Resource

	pod *v1.Pod
}

// List returns a collection of containers
func (c *Container) List(ctx context.Context) ([]runtime.Object, error) {
	c.pod = nil
	path, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, fmt.Errorf("no context path for %q", c.gvr)
	}
	ns, _ := render.Namespaced(path)
	c.namespace = ns
	o, err := c.factory.Get("v1/pods", path, labels.Everything())
	if err != nil {
		return nil, err
	}

	var po v1.Pod
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &po)
	if err != nil {
		return nil, err
	}
	c.pod = &po

	res := make([]runtime.Object, 0, len(po.Spec.InitContainers)+len(po.Spec.Containers))
	for _, co := range po.Spec.InitContainers {
		res = append(res, ContainerRes{co})
	}
	for _, co := range po.Spec.Containers {
		res = append(res, ContainerRes{co})
	}

	return res, nil
}

// Hydrate returns a pod as container rows.
func (c *Container) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	mx := client.NewMetricsServer(c.factory.Client().(client.Connection))
	mmx, err := mx.FetchPodMetrics(c.namespace, c.pod.Name)
	if err != nil {
		log.Warn().Err(err).Msgf("No metrics found for pod %q:%q", c.namespace, c.pod.Name)
	}

	var index int
	for _, o := range oo {
		co := o.(ContainerRes)
		row, err := renderCoRow(co.Container.Name, index, coMetricsFor(co.Container, c.pod, mmx, true), re)
		if err != nil {
			return err
		}
		rr[index] = row
		index++
	}

	return nil
}

func renderCoRow(n string, index int, pmx *ContainerWithMetrics, re Renderer) (render.Row, error) {
	var row render.Row
	if err := re.Render(pmx, n, &row); err != nil {
		return render.Row{}, err
	}
	return row, nil
}

func coMetricsFor(co v1.Container, po *v1.Pod, mmx *mv1beta1.PodMetrics, isInit bool) *ContainerWithMetrics {
	return &ContainerWithMetrics{
		container: &co,
		status:    getContainerStatus(co.Name, po.Status),
		metrics:   containerMetrics(co.Name, mmx),
		isInit:    isInit,
		age:       po.ObjectMeta.CreationTimestamp,
	}
}

func containerMetrics(n string, mx runtime.Object) *mv1beta1.ContainerMetrics {
	pmx := mx.(*mv1beta1.PodMetrics)
	for _, m := range pmx.Containers {
		if m.Name == n {
			return &m
		}
	}
	return nil
}

// ----------------------------------------------------------------------------

func getContainerStatus(co string, status v1.PodStatus) *v1.ContainerStatus {
	for _, c := range status.ContainerStatuses {
		if c.Name == co {
			return &c
		}
	}

	for _, c := range status.InitContainerStatuses {
		if c.Name == co {
			return &c
		}
	}

	return nil
}

// ContainerWithMetrics represents a container and its metrics.
type ContainerWithMetrics struct {
	container *v1.Container
	status    *v1.ContainerStatus
	metrics   *mv1beta1.ContainerMetrics
	isInit    bool
	age       metav1.Time
}

func (c *ContainerWithMetrics) IsInit() bool {
	return c.isInit
}

func (c *ContainerWithMetrics) Container() *v1.Container {
	return c.container
}

func (c *ContainerWithMetrics) ContainerStatus() *v1.ContainerStatus {
	return c.status
}

// Metrics returns the metrics associated with the pod.
func (c *ContainerWithMetrics) Metrics() *mv1beta1.ContainerMetrics {
	return c.metrics
}

func (c *ContainerWithMetrics) Age() metav1.Time {
	return c.age
}

// ----------------------------------------------------------------------------

// ContainerRes represents a container K8s resource.
type ContainerRes struct {
	v1.Container
}

// GetObjectKind returns a schema object.
func (c ContainerRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (c ContainerRes) DeepCopyObject() runtime.Object {
	return c
}
