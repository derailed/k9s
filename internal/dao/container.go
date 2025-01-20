// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"
	"strconv"

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

const (
	initIDX = "I"
	mainIDX = "M"
	ephIDX  = "E"
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
	res := make([]runtime.Object, 0, len(po.Spec.InitContainers)+len(po.Spec.Containers)+len(po.Spec.EphemeralContainers))
	for i, co := range po.Spec.InitContainers {
		res = append(res, makeContainerRes(initIDX, i, co, po, cmx[co.Name]))
	}
	for i, co := range po.Spec.Containers {
		res = append(res, makeContainerRes(mainIDX, i, co, po, cmx[co.Name]))
	}
	for i, co := range po.Spec.EphemeralContainers {
		res = append(res, makeContainerRes(ephIDX, i, v1.Container(co.EphemeralContainerCommon), po, cmx[co.Name]))
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

func makeContainerRes(kind string, idx int, co v1.Container, po *v1.Pod, cmx *mv1beta1.ContainerMetrics) render.ContainerRes {
	return render.ContainerRes{
		Idx:       kind + strconv.Itoa(idx+1),
		Container: &co,
		Status:    getContainerStatus(kind, co.Name, po.Status),
		MX:        cmx,
		Age:       po.GetCreationTimestamp(),
	}
}

func getContainerStatus(kind string, name string, status v1.PodStatus) *v1.ContainerStatus {
	switch kind {
	case mainIDX:
		for _, s := range status.ContainerStatuses {
			if s.Name == name {
				return &s
			}
		}
	case initIDX:
		for _, s := range status.InitContainerStatuses {
			if s.Name == name {
				return &s
			}
		}
	case ephIDX:
		for _, s := range status.EphemeralContainerStatuses {
			if s.Name == name {
				return &s
			}
		}
	}

	return nil
}

func (c *Container) fetchPod(fqn string) (*v1.Pod, error) {
	o, err := c.getFactory().Get("v1/pods", fqn, true, labels.Everything())
	if err != nil {
		return nil, fmt.Errorf("failed to locate pod %q: %w", fqn, err)
	}
	var po v1.Pod
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &po)
	return &po, err
}
