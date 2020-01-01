package model

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

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
	o, err := c.factory.Get("v1/pods", path, true, labels.Everything())
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
	mx := client.NewMetricsServer(c.factory.Client())
	var pmx *mv1beta1.PodMetrics
	if c.factory.Client() != nil {
		var err error
		pmx, err = mx.FetchPodMetrics(c.namespace, c.pod.Name)
		if err != nil {
			log.Warn().Err(err).Msgf("No metrics found for pod %q:%q", c.namespace, c.pod.Name)
		}
	}

	for _, co := range po.Spec.InitContainers {
		res = append(res, makeContainerRes(co, po, pmx, true))
	}
	for _, co := range po.Spec.Containers {
		res = append(res, makeContainerRes(co, po, pmx, false))
	}

	return res, nil
}

// ----------------------------------------------------------------------------
// Helpers...

func makeContainerRes(co v1.Container, po v1.Pod, pmx *mv1beta1.PodMetrics, isInit bool) render.ContainerRes {
	cmx, err := containerMetrics(co.Name, pmx)
	if err != nil {
		log.Warn().Err(err).Msgf("Container metrics for %s", co.Name)
	}

	return render.ContainerRes{
		Container: co,
		Status:    getContainerStatus(co.Name, po.Status),
		Metrics:   cmx,
		IsInit:    isInit,
		Age:       po.ObjectMeta.CreationTimestamp,
	}
}

func containerMetrics(n string, mx runtime.Object) (*mv1beta1.ContainerMetrics, error) {
	pmx, ok := mx.(*mv1beta1.PodMetrics)
	if !ok {
		return nil, fmt.Errorf("expecting podmetrics but got `%T", mx)
	}
	if pmx == nil {
		return nil, fmt.Errorf("no metrics for container %s", n)
	}
	for _, m := range pmx.Containers {
		if m.Name == n {
			return &m, nil
		}
	}
	return nil, nil
}

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
