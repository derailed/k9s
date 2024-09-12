// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"fmt"
	"github.com/fvbommel/sortorder"
	"sort"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestContainerRes_IndexLabel(t *testing.T) {
	cresInit0 := render.ContainerRes{
		Index:  0,
		IsInit: true,
	}
	cresInit1 := render.ContainerRes{
		Index:  1,
		IsInit: true,
	}
	cresReg0 := render.ContainerRes{
		Index:  0,
		IsInit: false,
	}
	cresReg1 := render.ContainerRes{
		Index:  1,
		IsInit: false,
	}

	assert.Equal(t, cresInit0.IndexLabel(), "ⓘ 0")
	assert.Equal(t, cresInit1.IndexLabel(), "ⓘ 1")
	assert.Equal(t, cresReg0.IndexLabel(), "⠀ 0")
	assert.Equal(t, cresReg1.IndexLabel(), "⠀ 1")

	assert.True(t, sort.IsSorted(sortorder.Natural{
		cresInit0.IndexLabel(),
		cresInit1.IndexLabel(),
		cresReg0.IndexLabel(),
		cresReg1.IndexLabel(),
	}))
}

func TestContainer(t *testing.T) {
	var c render.Container

	cres := makeContainerRes(
		makeContainer(),
		makeContainerStatus(),
		makeContainerMetrics(),
		false,
		makeAge(),
	)
	var r model1.Row
	assert.Nil(t, c.Render(cres, "blee", &r))
	assert.Equal(t, "⠀ 0", r.ID)
	assert.Equal(t, model1.Fields{
		"⠀ 0",
		"fred",
		"●",
		"img",
		"false",
		"Running",
		"0",
		"off:off",
		"10",
		"20",
		"20:20",
		"100:100",
		"50",
		"50",
		"20",
		"20",
		"",
		"container is not ready",
	},
		r.Fields[:len(r.Fields)-1],
	)
}

func TestInitContainer(t *testing.T) {
	var c render.Container

	cres := makeContainerRes(
		makeContainer(),
		makeContainerStatus(),
		makeContainerMetrics(),
		true,
		makeAge(),
	)
	var r model1.Row
	assert.Nil(t, c.Render(cres, "blee", &r))
	assert.Equal(t, "ⓘ 0", r.ID)
	assert.Equal(t, model1.Fields{
		"ⓘ 0",
		"fred",
		"●",
		"img",
		"false",
		"Running",
		"0",
		"off:off",
		"10",
		"20",
		"20:20",
		"100:100",
		"50",
		"50",
		"20",
		"20",
		"",
		"container is not ready",
	},
		r.Fields[:len(r.Fields)-1],
	)
}

func BenchmarkContainerRender(b *testing.B) {
	var c render.Container

	cres := makeContainerRes(
		makeContainer(),
		makeContainerStatus(),
		makeContainerMetrics(),
		false,
		makeAge(),
	)
	var r model1.Row

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = c.Render(cres, "blee", &r)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeContainerRes(container *v1.Container, status *v1.ContainerStatus, cmx *mv1beta1.ContainerMetrics, isInit bool, age metav1.Time) render.ContainerRes {
	po := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			CreationTimestamp: age,
		},
	}

	if isInit {
		po.Spec.InitContainers = []v1.Container{*container}
		po.Status.InitContainerStatuses = []v1.ContainerStatus{*status}
	} else {
		po.Spec.Containers = []v1.Container{*container}
		po.Status.ContainerStatuses = []v1.ContainerStatus{*status}
	}

	cr := render.MakeContainerRes(
		po,
		isInit,
		0,
	)
	cr.MX = cmx
	return cr
}

func toQty(s string) resource.Quantity {
	q, _ := resource.ParseQuantity(s)
	return q
}

func makeContainerMetrics() *mv1beta1.ContainerMetrics {
	return &mv1beta1.ContainerMetrics{
		Name: "fred",
		Usage: v1.ResourceList{
			v1.ResourceCPU:    toQty("10m"),
			v1.ResourceMemory: toQty("20Mi"),
		},
	}
}

func makeAge() metav1.Time {
	return metav1.Time{Time: testTime()}
}

func makeContainer() *v1.Container {
	return &v1.Container{
		Name:  "fred",
		Image: "img",
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:    toQty("20m"),
				v1.ResourceMemory: toQty("100Mi"),
			},
		},
		Env: []v1.EnvVar{
			{
				Name:  "fred",
				Value: "1",
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{Key: "blee"},
				},
			},
		},
	}
}

func makeContainerStatus() *v1.ContainerStatus {
	return &v1.ContainerStatus{
		Name:         "fred",
		State:        v1.ContainerState{Running: &v1.ContainerStateRunning{}},
		RestartCount: 0,
	}
}

func testTime() time.Time {
	t, err := time.Parse(time.RFC3339, "2018-12-14T10:36:43.326972-07:00")
	if err != nil {
		fmt.Println("TestTime Failed", err)
	}
	return t
}
