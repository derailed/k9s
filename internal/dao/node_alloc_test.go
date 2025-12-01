// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"testing"

	"github.com/derailed/k9s/internal/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNodeAllocCountPods(t *testing.T) {
	uu := map[string]struct {
		pods     []runtime.Object
		nodeName string
		expected int
		err      bool
	}{
		"no_pods": {
			pods:     []runtime.Object{},
			nodeName: "node1",
			expected: 0,
		},
		"single_pod": {
			pods: []runtime.Object{
				makePodUnstructured("pod1", "node1", v1.PodRunning),
			},
			nodeName: "node1",
			expected: 1,
		},
		"multiple_pods_same_node": {
			pods: []runtime.Object{
				makePodUnstructured("pod1", "node1", v1.PodRunning),
				makePodUnstructured("pod2", "node1", v1.PodRunning),
				makePodUnstructured("pod3", "node1", v1.PodRunning),
			},
			nodeName: "node1",
			expected: 3,
		},
		"pods_different_nodes": {
			pods: []runtime.Object{
				makePodUnstructured("pod1", "node1", v1.PodRunning),
				makePodUnstructured("pod2", "node2", v1.PodRunning),
				makePodUnstructured("pod3", "node1", v1.PodRunning),
			},
			nodeName: "node1",
			expected: 2,
		},
		"invalid_object": {
			pods: []runtime.Object{
				&v1.Pod{},
			},
			nodeName: "node1",
			err:      true,
		},
	}

	var na dao.NodeAlloc
	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			count, err := na.CountPods(u.pods, u.nodeName)
			if u.err {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, u.expected, count)
			}
		})
	}
}

func TestNodeAllocCalculateRequestedResources(t *testing.T) {
	uu := map[string]struct {
		pods        []runtime.Object
		nodeName    string
		expectedCPU int64
		expectedMem int64
		expectedErr bool
	}{
		"no_pods": {
			pods:        []runtime.Object{},
			nodeName:    "node1",
			expectedCPU: 0,
			expectedMem: 0,
		},
		"single_pod_with_requests": {
			pods: []runtime.Object{
				makePodUnstructuredWithResources("pod1", "node1", v1.PodRunning, "100m", "128Mi", "", ""),
			},
			nodeName:    "node1",
			expectedCPU: 100,
			expectedMem: 134217728,
		},
		"multiple_pods": {
			pods: []runtime.Object{
				makePodUnstructuredWithResources("pod1", "node1", v1.PodRunning, "200m", "256Mi", "", ""),
				makePodUnstructuredWithResources("pod2", "node1", v1.PodRunning, "300m", "512Mi", "", ""),
			},
			nodeName:    "node1",
			expectedCPU: 500,
			expectedMem: 805306368,
		},
		"pods_different_nodes": {
			pods: []runtime.Object{
				makePodUnstructuredWithResources("pod1", "node1", v1.PodRunning, "100m", "128Mi", "", ""),
				makePodUnstructuredWithResources("pod2", "node2", v1.PodRunning, "200m", "256Mi", "", ""),
			},
			nodeName:    "node1",
			expectedCPU: 100,
			expectedMem: 134217728,
		},
		"terminated_pods_skipped": {
			pods: []runtime.Object{
				makePodUnstructuredWithResources("pod1", "node1", v1.PodRunning, "100m", "128Mi", "", ""),
				makePodUnstructuredWithResources("pod2", "node1", v1.PodSucceeded, "200m", "256Mi", "", ""),
				makePodUnstructuredWithResources("pod3", "node1", v1.PodFailed, "300m", "512Mi", "", ""),
			},
			nodeName:    "node1",
			expectedCPU: 100,
			expectedMem: 134217728,
		},
		"pod_with_limits_only": {
			pods: []runtime.Object{
				makePodUnstructuredWithResources("pod1", "node1", v1.PodRunning, "", "", "100m", "128Mi"),
			},
			nodeName:    "node1",
			expectedCPU: 100,
			expectedMem: 134217728,
		},
		"pod_with_init_container": {
			pods: []runtime.Object{
				makePodUnstructuredWithInitContainer("pod1", "node1", v1.PodRunning, "100m", "128Mi", "500m", "512Mi"),
			},
			nodeName:    "node1",
			expectedCPU: 600,
			expectedMem: 671088640,
		},
		"pod_with_init_container_max": {
			pods: []runtime.Object{
				makePodUnstructuredWithInitContainer("pod1", "node1", v1.PodRunning, "500m", "512Mi", "100m", "128Mi"),
			},
			nodeName:    "node1",
			expectedCPU: 600,
			expectedMem: 671088640,
		},
		"pod_no_resources": {
			pods: []runtime.Object{
				makePodUnstructuredWithResources("pod1", "node1", v1.PodRunning, "", "", "", ""),
			},
			nodeName:    "node1",
			expectedCPU: 0,
			expectedMem: 0,
		},
	}

	var na dao.NodeAlloc
	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			cpu, mem, err := na.CalculateRequestedResources(u.pods, u.nodeName)
			if u.expectedErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, u.expectedCPU, cpu)
				assert.Equal(t, u.expectedMem, mem)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makePodUnstructured(name, nodeName string, phase v1.PodPhase) *unstructured.Unstructured {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			NodeName: nodeName,
		},
		Status: v1.PodStatus{
			Phase: phase,
		},
	}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
	if err != nil {
		panic(err)
	}

	return &unstructured.Unstructured{Object: obj}
}

func makePodUnstructuredWithResources(name, nodeName string, phase v1.PodPhase, cpuReq, memReq, cpuLim, memLim string) *unstructured.Unstructured {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			NodeName: nodeName,
			Containers: []v1.Container{
				{
					Name:      "container1",
					Resources: v1.ResourceRequirements{},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: phase,
		},
	}

	if cpuReq != "" || memReq != "" {
		pod.Spec.Containers[0].Resources.Requests = makeResourceList(cpuReq, memReq)
	}
	if cpuLim != "" || memLim != "" {
		pod.Spec.Containers[0].Resources.Limits = makeResourceList(cpuLim, memLim)
	}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
	if err != nil {
		panic(err)
	}

	return &unstructured.Unstructured{Object: obj}
}

func makePodUnstructuredWithInitContainer(name, nodeName string, phase v1.PodPhase, cpuReq, memReq, initCPUReq, initMemReq string) *unstructured.Unstructured {
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			NodeName: nodeName,
			InitContainers: []v1.Container{
				{
					Name: "init-container",
					Resources: v1.ResourceRequirements{
						Requests: makeResourceList(initCPUReq, initMemReq),
					},
				},
			},
			Containers: []v1.Container{
				{
					Name: "container1",
					Resources: v1.ResourceRequirements{
						Requests: makeResourceList(cpuReq, memReq),
					},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: phase,
		},
	}

	obj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(pod)
	if err != nil {
		panic(err)
	}

	return &unstructured.Unstructured{Object: obj}
}

func makeResourceList(cpu, mem string) v1.ResourceList {
	rl := v1.ResourceList{}
	if cpu != "" {
		rl[v1.ResourceCPU] = resource.MustParse(cpu)
	}
	if mem != "" {
		rl[v1.ResourceMemory] = resource.MustParse(mem)
	}
	return rl
}
