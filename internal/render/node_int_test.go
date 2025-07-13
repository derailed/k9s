package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func Test_gatherNodeMX(t *testing.T) {
	uu := map[string]struct {
		node   v1.Node
		nMX    *mv1beta1.NodeMetrics
		ec, ea metric
	}{
		"empty": {},

		"nvidia": {
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nvidia",
				},
				Status: v1.NodeStatus{
					Capacity: v1.ResourceList{
						v1.ResourceCPU:                    resource.MustParse("3"),
						v1.ResourceMemory:                 resource.MustParse("4Gi"),
						v1.ResourceName("nvidia.com/gpu"): resource.MustParse("2"),
					},
					Allocatable: v1.ResourceList{
						v1.ResourceCPU:                    resource.MustParse("8"),
						v1.ResourceMemory:                 resource.MustParse("8Gi"),
						v1.ResourceName("nvidia.com/gpu"): resource.MustParse("4"),
					},
				},
			},
			nMX: &mv1beta1.NodeMetrics{
				ObjectMeta: metav1.ObjectMeta{
					Name: "nvidia",
				},
				Usage: v1.ResourceList{
					v1.ResourceCPU:                    resource.MustParse("3"),
					v1.ResourceMemory:                 resource.MustParse("4Gi"),
					v1.ResourceName("nvidia.com/gpu"): resource.MustParse("2"),
				},
			},
			ea: metric{
				cpu: 8000,
				mem: 8589934592,
				gpu: 4,
			},
			ec: metric{
				cpu: 3000,
				mem: 4294967296,
				gpu: 2,
			},
		},

		"intel": {
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "intel",
				},
				Status: v1.NodeStatus{
					Capacity: v1.ResourceList{
						v1.ResourceCPU:                        resource.MustParse("3"),
						v1.ResourceMemory:                     resource.MustParse("4Gi"),
						v1.ResourceName("gpu.intel.com/i915"): resource.MustParse("2"),
					},
					Allocatable: v1.ResourceList{
						v1.ResourceCPU:                        resource.MustParse("8"),
						v1.ResourceMemory:                     resource.MustParse("8Gi"),
						v1.ResourceName("gpu.intel.com/i915"): resource.MustParse("4"),
					},
				},
			},
			ea: metric{
				cpu: 8000,
				mem: 8589934592,
				gpu: 4,
			},
			ec: metric{
				cpu: 0,
				mem: 0,
				gpu: 2,
			},
		},

		"unknown-vendor": {
			node: v1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Name: "amd",
				},
				Status: v1.NodeStatus{
					Capacity: v1.ResourceList{
						v1.ResourceCPU:              resource.MustParse("3"),
						v1.ResourceMemory:           resource.MustParse("4Gi"),
						v1.ResourceName("bozo/gpu"): resource.MustParse("2"),
					},
					Allocatable: v1.ResourceList{
						v1.ResourceCPU:              resource.MustParse("8"),
						v1.ResourceMemory:           resource.MustParse("8Gi"),
						v1.ResourceName("bozo/gpu"): resource.MustParse("4"),
					},
				},
			},
			ea: metric{
				cpu: 8000,
				mem: 8589934592,
				gpu: 0,
			},
			ec: metric{
				gpu: 0,
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			c, a := gatherNodeMX(&u.node, u.nMX)
			assert.Equal(t, u.ec, c)
			assert.Equal(t, u.ea, a)
		})
	}
}
