// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/slogs"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*NodeAlloc)(nil)
)

// NodeAlloc represents a node allocation model that always calculates allocated resources.
type NodeAlloc struct {
	Resource
}

// List returns a collection of node resources with allocated CPU and memory.
// This implementation always enables pod counting to calculate allocated resources.
func (n *NodeAlloc) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	// NodeAllocGVR is a k9s alias, not a real Kubernetes resource.
	// We need to list nodes using NodeGVR instead.
	lsel := labels.Everything()
	if sel, ok := ctx.Value(internal.KeyLabels).(labels.Selector); ok {
		lsel = sel
	}
	oo, err := n.getFactory().List(client.NodeGVR, ns, false, lsel)
	if err != nil {
		return oo, err
	}

	var nmx client.NodesMetricsMap
	if withMx, ok := ctx.Value(internal.KeyWithMetrics).(bool); withMx || !ok {
		nmx, _ = client.DialMetrics(n.Client()).FetchNodesMetricsMap(ctx)
	}

	// Always enable pod counting for node allocation view
	var pods []runtime.Object
	pods, err = n.getFactory().List(client.PodGVR, client.BlankNamespace, false, labels.Everything())
	if err != nil {
		slog.Error("Unable to list pods", slogs.Error, err)
	}

	res := make([]runtime.Object, 0, len(oo))
	for _, o := range oo {
		u, ok := o.(*unstructured.Unstructured)
		if !ok {
			return res, fmt.Errorf("expecting *unstructured.Unstructured but got `%T", o)
		}

		fqn := extractFQN(o)
		_, name := client.Namespaced(fqn)
		podCount := -1
		var requestedCPU, requestedMemory int64 = -1, -1

		// Always calculate pod count and requested resources
		podCount, err = n.CountPods(pods, name)
		if err != nil {
			slog.Error("Unable to get pods count",
				slogs.ResName, name,
				slogs.Error, err,
			)
		}
		requestedCPU, requestedMemory, err = n.CalculateRequestedResources(pods, name)
		if err != nil {
			slog.Error("Unable to calculate requested resources",
				slogs.ResName, name,
				slogs.Error, err,
			)
		}

		res = append(res, &render.NodeWithMetrics{
			Raw:             u,
			MX:              nmx[name],
			PodCount:        podCount,
			RequestedCPU:    requestedCPU,
			RequestedMemory: requestedMemory,
		})
	}

	return res, nil
}

// CountPods counts the pods scheduled on a given node.
func (*NodeAlloc) CountPods(oo []runtime.Object, nodeName string) (int, error) {
	var count int
	for _, o := range oo {
		u, ok := o.(*unstructured.Unstructured)
		if !ok {
			return count, fmt.Errorf("expecting *Unstructured but got `%T", o)
		}
		spec, ok := u.Object["spec"].(map[string]any)
		if !ok {
			return count, fmt.Errorf("expecting spec interface map but got `%T", o)
		}
		if node, ok := spec["nodeName"]; ok && node == nodeName {
			count++
		}
	}

	return count, nil
}

// CalculateRequestedResources calculates the sum of CPU and memory requests
// from all non-terminated pods on a given node.
func (n *NodeAlloc) CalculateRequestedResources(oo []runtime.Object, nodeName string) (cpu int64, memory int64, err error) {
	cpuQ := resource.NewQuantity(0, resource.DecimalSI)
	memQ := resource.NewQuantity(0, resource.BinarySI)

	for _, o := range oo {
		u, ok := o.(*unstructured.Unstructured)
		if !ok {
			continue
		}
		spec, ok := u.Object["spec"].(map[string]any)
		if !ok {
			continue
		}
		if node, ok := spec["nodeName"]; !ok || node != nodeName {
			continue
		}

		var pod v1.Pod
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &pod); err != nil {
			continue
		}

		// Skip terminated pods (Succeeded or Failed)
		if pod.Status.Phase == v1.PodSucceeded || pod.Status.Phase == v1.PodFailed {
			continue
		}

		var initCPU, initMem *resource.Quantity
		for i := range pod.Spec.InitContainers {
			co := pod.Spec.InitContainers[i]
			req := co.Resources.Requests
			if len(req) == 0 {
				req = co.Resources.Limits
			}
			if len(req) == 0 {
				continue
			}

			if q := req.Cpu(); q != nil {
				if initCPU == nil || q.Cmp(*initCPU) > 0 {
					initCPU = q
				}
			}
			if q := req.Memory(); q != nil {
				if initMem == nil || q.Cmp(*initMem) > 0 {
					initMem = q
				}
			}
		}

		podCPUQ := resource.NewQuantity(0, resource.DecimalSI)
		podMemQ := resource.NewQuantity(0, resource.BinarySI)

		if initCPU != nil {
			podCPUQ.Add(*initCPU)
		}
		if initMem != nil {
			podMemQ.Add(*initMem)
		}

		for i := range pod.Spec.Containers {
			co := pod.Spec.Containers[i]
			req := co.Resources.Requests
			if len(req) == 0 {
				req = co.Resources.Limits
			}
			if len(req) == 0 {
				continue
			}

			if q := req.Cpu(); q != nil {
				podCPUQ.Add(*q)
			}
			if q := req.Memory(); q != nil {
				podMemQ.Add(*q)
			}
		}

		slog.Debug("Calculating requested resources for non-terminated pod",
			slogs.ResName, pod.Name,
			slogs.Namespace, pod.Namespace,
			slogs.PodPhase, pod.Status.Phase,
			"node", nodeName,
			"cpu-requested-m", podCPUQ.MilliValue(),
			"memory-requested-bytes", podMemQ.Value(),
		)

		cpuQ.Add(*podCPUQ)
		memQ.Add(*podMemQ)
	}

	return cpuQ.MilliValue(), memQ.Value(), nil
}
