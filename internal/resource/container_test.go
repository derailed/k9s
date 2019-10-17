package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestProbe(t *testing.T) {
	uu := map[string]struct {
		probe *v1.Probe
		e     string
	}{
		"defined":   {&v1.Probe{}, "on"},
		"undefined": {nil, "off"},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, probe(u.probe))
		})
	}
}

func TestAsMi(t *testing.T) {
	uu := map[string]struct {
		mem int64
		e   float64
	}{
		"zero": {0, 0},
		"1Mb":  {1024 * 1024, 1.048576e+06},
		"10Mb": {10 * 1024 * 1024, 1.048576e+07},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, asMi(u.mem))
		})
	}
}

func TestToRes(t *testing.T) {
	uu := map[string]struct {
		res        v1.ResourceList
		ecpu, emem string
	}{
		"cool": {v1.ResourceList{
			v1.ResourceCPU:    toQty("10m"),
			v1.ResourceMemory: toQty("20Mi"),
		},
			"10", "20"},
		"noRes": {v1.ResourceList{},
			"0", "0"},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			cpu, mem := toRes(u.res)
			assert.Equal(t, u.ecpu, cpu)
			assert.Equal(t, u.emem, mem)
		})
	}
}

func TestToState(t *testing.T) {
	uu := map[string]struct {
		state v1.ContainerState
		e     string
	}{
		"empty": {v1.ContainerState{},
			MissingValue},
		"running": {
			v1.ContainerState{Running: &v1.ContainerStateRunning{}},
			"Running",
		},
		"waiting": {
			v1.ContainerState{Waiting: &v1.ContainerStateWaiting{}},
			"Waiting",
		},
		"waitingReason": {
			v1.ContainerState{Waiting: &v1.ContainerStateWaiting{Reason: "blee"}},
			"blee",
		},
		"terminating": {
			v1.ContainerState{Terminated: &v1.ContainerStateTerminated{}},
			"Terminating",
		},
		"terminatedReason": {
			v1.ContainerState{Terminated: &v1.ContainerStateTerminated{Reason: "blee"}},
			"blee",
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, toState(u.state))
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func toQty(s string) resource.Quantity {
	q, _ := resource.ParseQuantity(s)

	return q
}
