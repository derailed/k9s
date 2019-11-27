package model

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var _ render.ContainerWithMetrics = &ContainerWithMetrics{}

// Container represents a container model.
type Container struct {
	*Resource
}

// NewContainer returns a new container model
func NewContainer() *Container {
	return &Container{
		Resource: NewResource(),
	}
}

// List returns a collection of containers
func (c *Container) List(sel string) ([]runtime.Object, error) {
	ns, n := render.Namespaced(sel)
	c.namespace = ns
	o, err := c.factory.Get(ns, "v1/pods", n, labels.Everything())
	if err != nil {
		return nil, err
	}

	var po v1.Pod
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &po)
	if err != nil {
		return nil, err
	}

	res := make([]runtime.Object, 1, len(po.Spec.InitContainers)+len(po.Spec.Containers))
	res[0] = &po
	return res, nil
}

// Hydrate returns a pod as container rows.
func (c *Container) Hydrate(cc []runtime.Object, rr render.Rows, re Renderer) error {
	po := cc[0].(*v1.Pod)
	mx := k8s.NewMetricsServer(c.factory.Client().(k8s.Connection))
	mmx, err := mx.FetchPodMetrics(c.namespace, po.Name)
	if err != nil {
		return err
	}

	var index int
	size := len(re.Header(c.namespace))
	for _, co := range po.Spec.InitContainers {
		row, err := renderCoRow(co.Name, index, size, coMetricsFor(co, po, mmx, true), re)
		if err != nil {
			return err
		}
		rr[index] = row
		log.Debug().Msgf("Init Containers %#v", rr[index])
		index++
	}
	for _, co := range po.Spec.Containers {
		row, err := renderCoRow(co.Name, index, size, coMetricsFor(co, po, mmx, false), re)
		if err != nil {
			return err
		}
		rr[index] = row
		log.Debug().Msgf("Containers %#v", row)
		index++
	}
	return nil
}

func renderCoRow(n string, index, size int, pmx *ContainerWithMetrics, re Renderer) (render.Row, error) {
	row := render.Row{Fields: make([]string, size)}
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
	log.Debug().Msgf("CO MX fo %s", n)
	for _, m := range pmx.Containers {
		log.Debug().Msgf("Container Metrics %#v", m)
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
