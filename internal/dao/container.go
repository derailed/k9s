package dao

import (
	"context"
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	restclient "k8s.io/client-go/rest"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var (
	_ Accessor = (*Container)(nil)
	_ Loggable = (*Container)(nil)
)

// Container represents a pod's container dao.
type Container struct {
	NonResource
}

// List returns a collection of containers.
func (c *Container) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	path, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, fmt.Errorf("no context path for %q", c.gvr)
	}
	ns, _ := render.Namespaced(path)
	o, err := c.Factory.Get("v1/pods", path, true, labels.Everything())
	if err != nil {
		return nil, err
	}

	var po v1.Pod
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &po)
	if err != nil {
		return nil, err
	}
	res := make([]runtime.Object, 0, len(po.Spec.InitContainers)+len(po.Spec.Containers))
	mx := client.NewMetricsServer(c.Client())
	var pmx *mv1beta1.PodMetrics
	if c.Client() != nil {
		var err error
		pmx, err = mx.FetchPodMetrics(ns, po.Name)
		if err != nil {
			log.Warn().Err(err).Msgf("No metrics found for pod %q:%q", ns, po.Name)
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

// TailLogs tails a given container logs
func (c *Container) TailLogs(ctx context.Context, logChan chan<- string, opts LogOptions) error {
	fac, ok := ctx.Value(internal.KeyFactory).(Factory)
	if !ok {
		return errors.New("Expecting an informer")
	}
	o, err := fac.Get("v1/pods", opts.Path, true, labels.Everything())
	if err != nil {
		return err
	}

	var po v1.Pod
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &po); err != nil {
		return err
	}

	return tailLogs(ctx, c, logChan, opts)
}

// Logs fetch container logs for a given pod and container.
func (c *Container) Logs(path string, opts *v1.PodLogOptions) (*restclient.Request, error) {
	ns, _ := client.Namespaced(path)
	auth, err := c.Client().CanI(ns, "v1/pods:log", []string{"get"})
	if !auth || err != nil {
		return nil, err
	}

	ns, n := client.Namespaced(path)
	return c.Client().DialOrDie().CoreV1().Pods(ns).GetLogs(n, opts), nil
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
