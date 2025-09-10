// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func Test_gatherContainerMX(t *testing.T) {
	uu := map[string]struct {
		container v1.Container
		mx        *mv1beta1.ContainerMetrics
		c, r      metric
	}{
		"empty": {},

		"amd-request": {
			container: v1.Container{
				Name:  "fred",
				Image: "img",
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("10m"),
						v1.ResourceMemory: resource.MustParse("20Mi"),
						"nvidia.com/gpu":  resource.MustParse("1"),
					},
				},
			},
			mx: &mv1beta1.ContainerMetrics{
				Name: "fred",
				Usage: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("10m"),
					v1.ResourceMemory: resource.MustParse("20Mi"),
				},
			},
			c: metric{
				cpu: 10,
				mem: 20971520,
			},
			r: metric{
				cpu: 10,
				gpu: 1,
				mem: 20971520,
			},
		},

		"amd-both": {
			container: v1.Container{
				Name:  "fred",
				Image: "img",
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("10m"),
						v1.ResourceMemory: resource.MustParse("20Mi"),
						"nvidia.com/gpu":  resource.MustParse("1"),
					},
					Limits: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("50m"),
						v1.ResourceMemory: resource.MustParse("100Mi"),
						"nvidia.com/gpu":  resource.MustParse("2"),
					},
				},
			},
			mx: &mv1beta1.ContainerMetrics{
				Name: "fred",
				Usage: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("10m"),
					v1.ResourceMemory: resource.MustParse("20Mi"),
				},
			},
			c: metric{
				cpu: 10,
				mem: 20971520,
			},
			r: metric{
				cpu:  10,
				gpu:  1,
				mem:  20971520,
				lcpu: 50,
				lgpu: 2,
				lmem: 104857600,
			},
		},

		"amd-limits": {
			container: v1.Container{
				Name:  "fred",
				Image: "img",
				Resources: v1.ResourceRequirements{
					Limits: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("50m"),
						v1.ResourceMemory: resource.MustParse("100Mi"),
						"nvidia.com/gpu":  resource.MustParse("2"),
					},
				},
			},
			mx: &mv1beta1.ContainerMetrics{
				Name: "fred",
				Usage: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("10m"),
					v1.ResourceMemory: resource.MustParse("20Mi"),
				},
			},
			c: metric{
				cpu: 10,
				mem: 20971520,
			},
			r: metric{
				cpu:  50,
				gpu:  2,
				mem:  104857600,
				lcpu: 50,
				lgpu: 2,
				lmem: 104857600,
			},
		},

		"amd-no-mx": {
			container: v1.Container{
				Name:  "fred",
				Image: "img",
				Resources: v1.ResourceRequirements{
					Requests: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("10m"),
						v1.ResourceMemory: resource.MustParse("20Mi"),
						"nvidia.com/gpu":  resource.MustParse("1"),
					},
					Limits: v1.ResourceList{
						v1.ResourceCPU:    resource.MustParse("50m"),
						v1.ResourceMemory: resource.MustParse("100Mi"),
						"nvidia.com/gpu":  resource.MustParse("2"),
					},
				},
			},
			r: metric{
				cpu:  10,
				gpu:  1,
				mem:  20971520,
				lcpu: 50,
				lgpu: 2,
				lmem: 104857600,
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			c, r := gatherContainerMX(&u.container, u.mx)
			assert.Equal(t, u.c, c)
			assert.Equal(t, u.r, r)
		})
	}
}
