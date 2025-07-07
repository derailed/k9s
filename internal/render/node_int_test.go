package render

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func Test_gpuSpec(t *testing.T) {
	uu := map[string]struct {
		capacity    v1.ResourceList
		allocatable v1.ResourceList
		e           string
	}{
		"empty": {
			e: NAValue,
		},

		"nvidia": {
			capacity: v1.ResourceList{
				v1.ResourceName("nvidia.com/gpu"): resource.MustParse("2"),
			},
			allocatable: v1.ResourceList{
				v1.ResourceName("nvidia.com/gpu"): resource.MustParse("4"),
			},
			e: "2/4 (nvidia)",
		},

		"intel": {
			capacity: v1.ResourceList{
				v1.ResourceName("gpu.intel.com/i915"): resource.MustParse("2"),
			},
			allocatable: v1.ResourceList{
				v1.ResourceName("gpu.intel.com/i915"): resource.MustParse("4"),
			},
			e: "2/4 (intel)",
		},

		"amd": {
			capacity: v1.ResourceList{
				v1.ResourceName("amd.com/gpu"): resource.MustParse("2"),
			},
			allocatable: v1.ResourceList{
				v1.ResourceName("amd.com/gpu"): resource.MustParse("4"),
			},
			e: "2/4 (amd)",
		},

		"toast-cap": {
			capacity: v1.ResourceList{
				v1.ResourceName("gpu.intel.com/iBOZO"): resource.MustParse("2"),
			},
			allocatable: v1.ResourceList{
				v1.ResourceName("gpu.intel.com/i915"): resource.MustParse("4"),
			},
			e: NAValue,
		},

		"toast-alloc": {
			capacity: v1.ResourceList{
				v1.ResourceName("gpu.intel.com/i915"): resource.MustParse("2"),
			},
			allocatable: v1.ResourceList{
				v1.ResourceName("gpu.intel.com/iBOZO"): resource.MustParse("4"),
			},
			e: NAValue,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			var n Node
			assert.Equal(t, u.e, n.gpuSpec(u.capacity, u.allocatable))
		})
	}
}
