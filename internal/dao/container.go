// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
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
	fqn, ok := ctx.Value(internal.KeyPath).(string)
	if !ok {
		return nil, fmt.Errorf("no context path for %q", c.gvr)
	}

	var (
		cmx client.ContainersMetrics
		err error
	)
	if withMx, ok := ctx.Value(internal.KeyWithMetrics).(bool); ok && withMx {
		cmx, _ = client.DialMetrics(c.Client()).FetchContainersMetrics(ctx, fqn)
	}

	po, err := c.fetchPod(fqn)
	if err != nil {
		return nil, err
	}

	res := make([]runtime.Object, 0, len(po.Spec.InitContainers)+len(po.Spec.EphemeralContainers)+len(po.Spec.Containers))
	containerTypes := map[string][]v1.Container{
		"I": po.Spec.InitContainers,
		"E": convertEphemeralContainersToContainers(po.Spec.EphemeralContainers),
		"M": po.Spec.Containers,
	}
	for prefix, containers := range containerTypes {
		for i, co := range containers {
			res = append(res, makeContainerRes(fmt.Sprintf("%s%d", prefix, i+1), co, po, cmx[co.Name]))
		}
	}
	return res, nil
}

// TailLogs tails a given container logs.
func (c *Container) TailLogs(ctx context.Context, opts *LogOptions) ([]LogChan, error) {
	po := Pod{}
	po.Init(c.Factory, client.NewGVR("v1/pods"))

	return po.TailLogs(ctx, opts)
}

// ----------------------------------------------------------------------------
// Helpers...

func makeContainerRes(index string, co v1.Container, po *v1.Pod, cmx *mv1beta1.ContainerMetrics) render.ContainerRes {
	return render.ContainerRes{
		Index:     index,
		Container: &co,
		Status:    getContainerStatus(co.Name, po.Status),
		MX:        cmx,
		Age:       po.GetCreationTimestamp(),
	}
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

func (c *Container) fetchPod(fqn string) (*v1.Pod, error) {
	o, err := c.getFactory().Get("v1/pods", fqn, true, labels.Everything())
	if err != nil {
		return nil, err
	}
	var po v1.Pod
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &po)
	return &po, err
}

// convertEphemeralContainersToContainers reduces EphemeralContainers to common fields with Containers
func convertEphemeralContainersToContainers(ecos []v1.EphemeralContainer) []v1.Container {
	cos := make([]v1.Container, len(ecos))
	for i, eco := range ecos {
		cos[i] = v1.Container(eco.EphemeralContainerCommon)
	}
	return cos
}
