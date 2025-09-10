// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	res "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func Test_checkInitContainerStatus(t *testing.T) {
	trueVal := true
	uu := map[string]struct {
		status       v1.ContainerStatus
		e            string
		count, total int
		restart      bool
	}{
		"none": {
			e: "Init:0/0",
		},

		"restart": {
			status: v1.ContainerStatus{
				Name:    "ic1",
				Started: &trueVal,
				State:   v1.ContainerState{},
			},
			restart: true,
			e:       "Init:0/0",
		},

		"no-restart": {
			status: v1.ContainerStatus{
				Name:    "ic1",
				Started: &trueVal,
				State:   v1.ContainerState{},
			},
			e: "Init:0/0",
		},

		"terminated-reason": {
			status: v1.ContainerStatus{
				Name: "ic1",
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						ExitCode: 1,
						Reason:   "blah",
					},
				},
			},
			e: "Init:blah",
		},

		"terminated-signal": {
			status: v1.ContainerStatus{
				Name: "ic1",
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						ExitCode: 1,
						Signal:   9,
					},
				},
			},
			e: "Init:Signal:9",
		},

		"terminated-code": {
			status: v1.ContainerStatus{
				Name: "ic1",
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						ExitCode: 1,
					},
				},
			},
			e: "Init:ExitCode:1",
		},

		"terminated-restart": {
			status: v1.ContainerStatus{
				Name: "ic1",
				State: v1.ContainerState{
					Terminated: &v1.ContainerStateTerminated{
						Reason: "blah",
					},
				},
			},
		},

		"waiting": {
			status: v1.ContainerStatus{
				Name: "ic1",
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason: "blah",
					},
				},
			},
			e: "Init:blah",
		},

		"waiting-init": {
			status: v1.ContainerStatus{
				Name: "ic1",
				State: v1.ContainerState{
					Waiting: &v1.ContainerStateWaiting{
						Reason: "PodInitializing",
					},
				},
			},
			e: "Init:0/0",
		},

		"running": {
			status: v1.ContainerStatus{
				Name: "ic1",
				State: v1.ContainerState{
					Running: &v1.ContainerStateRunning{},
				},
			},
			e: "Init:0/0",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, checkInitContainerStatus(&u.status, u.count, u.total, u.restart))
		})
	}
}

func Test_containerPhase(t *testing.T) {
	uu := map[string]struct {
		status v1.PodStatus
		e      string
		ok     bool
	}{
		"none": {},

		"empty": {
			status: v1.PodStatus{
				Phase: PhaseUnknown,
			},
		},

		"waiting": {
			status: v1.PodStatus{
				Phase: PhaseUnknown,
				InitContainerStatuses: []v1.ContainerStatus{
					{
						Name: "ic1",
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name: "c1",
						State: v1.ContainerState{
							Waiting: &v1.ContainerStateWaiting{
								Reason: "waiting",
							},
						},
					},
				},
			},
			e: "waiting",
		},

		"terminated": {
			status: v1.PodStatus{
				Phase: PhaseUnknown,
				InitContainerStatuses: []v1.ContainerStatus{
					{
						Name: "ic1",
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name: "c1",
						State: v1.ContainerState{
							Terminated: &v1.ContainerStateTerminated{
								Reason: "done",
							},
						},
					},
				},
			},
			e: "done",
		},

		"terminated-sig": {
			status: v1.PodStatus{
				Phase: PhaseUnknown,
				InitContainerStatuses: []v1.ContainerStatus{
					{
						Name: "ic1",
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name: "c1",
						State: v1.ContainerState{
							Terminated: &v1.ContainerStateTerminated{
								Signal: 9,
							},
						},
					},
				},
			},
			e: "Signal:9",
		},

		"terminated-code": {
			status: v1.PodStatus{
				Phase: PhaseUnknown,
				InitContainerStatuses: []v1.ContainerStatus{
					{
						Name: "ic1",
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name: "c1",
						State: v1.ContainerState{
							Terminated: &v1.ContainerStateTerminated{
								ExitCode: 2,
							},
						},
					},
				},
			},
			e: "ExitCode:2",
		},

		"running": {
			status: v1.PodStatus{
				Phase: PhaseUnknown,
				InitContainerStatuses: []v1.ContainerStatus{
					{
						Name: "ic1",
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name:  "c1",
						Ready: true,
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
			},
			ok: true,
		},
	}

	var p Pod
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			s, ok := p.containerPhase(&u.status, "")
			assert.Equal(t, u.ok, ok)
			assert.Equal(t, u.e, s)
		})
	}
}

func Test_isSideCarContainer(t *testing.T) {
	always, never := v1.ContainerRestartPolicyAlways, v1.ContainerRestartPolicy("never")
	uu := map[string]struct {
		p *v1.ContainerRestartPolicy
		e bool
	}{
		"empty": {},

		"sidecar": {
			p: &always,
			e: true,
		},

		"no-sidecar": {
			p: &never,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, isSideCarContainer(u.p))
		})
	}
}

func Test_filterSidecarCO(t *testing.T) {
	always := v1.ContainerRestartPolicyAlways

	uu := map[string]struct {
		cc, ecc []v1.Container
	}{
		"empty": {
			cc:  []v1.Container{},
			ecc: []v1.Container{},
		},

		"restartable": {
			cc: []v1.Container{
				{
					Name:          "c1",
					RestartPolicy: &always,
				},
			},
			ecc: []v1.Container{
				{
					Name:          "c1",
					RestartPolicy: &always,
				},
			},
		},

		"not-restartable": {
			cc: []v1.Container{
				{
					Name: "c1",
				},
			},
			ecc: []v1.Container{},
		},

		"mixed": {
			cc: []v1.Container{
				{
					Name: "c1",
				},
				{
					Name:          "c2",
					RestartPolicy: &always,
				},
			},
			ecc: []v1.Container{
				{
					Name:          "c2",
					RestartPolicy: &always,
				},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.ecc, filterSidecarCO(u.cc))
		})
	}
}

func Test_lastRestart(t *testing.T) {
	uu := map[string]struct {
		containerStatuses []v1.ContainerStatus
		expected          metav1.Time
	}{
		"no-restarts": {
			containerStatuses: []v1.ContainerStatus{
				{
					Name:                 "c1",
					LastTerminationState: v1.ContainerState{},
				},
			},
			expected: metav1.Time{},
		},
		"single-container-restart": {
			containerStatuses: []v1.ContainerStatus{
				{
					Name: "c1",
					LastTerminationState: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							FinishedAt: metav1.Time{Time: testTime()},
						},
					},
				},
			},
			expected: metav1.Time{Time: testTime()},
		},
		"multiple-container-restarts": {
			containerStatuses: []v1.ContainerStatus{
				{
					Name: "c1",
					LastTerminationState: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							FinishedAt: metav1.Time{Time: testTime().Add(-1 * time.Hour)},
						},
					},
				},
				{
					Name: "c2",
					LastTerminationState: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							FinishedAt: metav1.Time{Time: testTime()},
						},
					},
				},
			},
			expected: metav1.Time{Time: testTime()},
		},
		"mixed-termination-states": {
			containerStatuses: []v1.ContainerStatus{
				{
					Name:                 "c1",
					LastTerminationState: v1.ContainerState{},
				},
				{
					Name: "c2",
					LastTerminationState: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{
							FinishedAt: metav1.Time{Time: testTime()},
						},
					},
				},
			},
			expected: metav1.Time{Time: testTime()},
		},
	}

	var p Pod
	for name, u := range uu {
		t.Run(name, func(t *testing.T) {
			_, _, _, lr := p.ContainerStats(u.containerStatuses)
			assert.Equal(t, u.expected, lr)
		})
	}
}

func Test_gatherPodMX(t *testing.T) {
	uu := map[string]struct {
		spec *v1.PodSpec
		mx   []mv1beta1.ContainerMetrics
		c, r metric
		perc string
	}{
		"single": {
			spec: &v1.PodSpec{
				Containers: []v1.Container{
					makeContainer("c1", false, "10m", "1Mi", "20m", "2Mi"),
				},
			},
			mx: []mv1beta1.ContainerMetrics{
				makeCoMX("c1", "1m", "22Mi"),
			},
			c: metric{
				cpu: 1,
				mem: 22 * client.MegaByte,
				gpu: 1,
			},
			r: metric{
				cpu:  10,
				mem:  1 * client.MegaByte,
				gpu:  1,
				lcpu: 20,
				lmem: 2 * client.MegaByte,
				lgpu: 1,
			},
			perc: "10",
		},

		"multi": {
			spec: &v1.PodSpec{
				Containers: []v1.Container{
					makeContainer("c1", false, "11m", "22Mi", "111m", "44Mi"),
					makeContainer("c2", false, "93m", "1402Mi", "0m", "2804Mi"),
					makeContainer("c3", false, "11m", "34Mi", "0m", "69Mi"),
				},
			},
			r: metric{
				cpu:  11 + 93 + 11,
				gpu:  1,
				mem:  (22 + 1402 + 34) * client.MegaByte,
				lcpu: 111 + 0 + 0,
				lgpu: 1,
				lmem: (44 + 2804 + 69) * client.MegaByte,
			},
			mx: []mv1beta1.ContainerMetrics{
				makeCoMX("c1", "1m", "22Mi"),
				makeCoMX("c2", "51m", "1275Mi"),
				makeCoMX("c3", "1m", "27Mi"),
			},
			c: metric{
				cpu: 1 + 51 + 1,
				gpu: 1,
				mem: (22 + 1275 + 27) * client.MegaByte,
			},
			perc: "46",
		},

		"sidecar": {
			spec: &v1.PodSpec{
				Containers: []v1.Container{
					makeContainer("c1", false, "11m", "22Mi", "111m", "44Mi"),
				},
				InitContainers: []v1.Container{
					makeContainer("c2", true, "93m", "1402Mi", "0m", "2804Mi"),
				},
			},
			r: metric{
				cpu:  11 + 93,
				gpu:  1,
				mem:  (22 + 1402) * client.MegaByte,
				lcpu: 111 + 0,
				lgpu: 1,
				lmem: (44 + 2804) * client.MegaByte,
			},
			mx: []mv1beta1.ContainerMetrics{
				makeCoMX("c1", "1m", "22Mi"),
				makeCoMX("c2", "51m", "1275Mi"),
			},
			c: metric{
				cpu: 1 + 51,
				gpu: 1,
				mem: (22 + 1275) * client.MegaByte,
			},
			perc: "50",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c, r := gatherPodMX(u.spec, u.mx)
			assert.Equal(t, u.c.cpu, c.cpu)
			assert.Equal(t, u.c.mem, c.mem)
			assert.Equal(t, u.c.lcpu, c.lcpu)
			assert.Equal(t, u.c.lmem, c.lmem)
			assert.Equal(t, u.c.lgpu, c.lgpu)

			assert.Equal(t, u.r.cpu, r.cpu)
			assert.Equal(t, u.r.mem, r.mem)
			assert.Equal(t, u.r.lcpu, r.lcpu)
			assert.Equal(t, u.r.lmem, r.lmem)
			assert.Equal(t, u.r.gpu, r.gpu)
			assert.Equal(t, u.r.lgpu, r.lgpu)

			assert.Equal(t, u.perc, client.ToPercentageStr(c.cpu, r.cpu))
		})
	}
}

func Test_podLimits(t *testing.T) {
	uu := map[string]struct {
		cc []v1.Container
		l  v1.ResourceList
	}{
		"plain": {
			cc: []v1.Container{
				makeContainer("c1", false, "10m", "1Mi", "20m", "2Mi"),
			},
			l: makeRes("20m", "2Mi"),
		},
		"multi-co": {
			cc: []v1.Container{
				makeContainer("c1", false, "10m", "1Mi", "20m", "2Mi"),
				makeContainer("c2", false, "10m", "1Mi", "40m", "4Mi"),
			},
			l: makeRes("60m", "6Mi"),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c, m, g := cosLimits(u.cc)
			assert.True(t, c.Equal(*u.l.Cpu()))
			assert.True(t, m.Equal(*u.l.Memory()))
			assert.True(t, g.Equal(*extractGPU(u.l)))
		})
	}
}

func Test_podRequests(t *testing.T) {
	uu := map[string]struct {
		cc []v1.Container
		e  v1.ResourceList
	}{
		"plain": {
			cc: []v1.Container{
				makeContainer("c1", false, "10m", "1Mi", "20m", "2Mi"),
			},
			e: makeRes("10m", "1Mi"),
		},

		"multi-co": {
			cc: []v1.Container{
				makeContainer("c1", false, "10m", "1Mi", "20m", "2Mi"),
				makeContainer("c2", false, "10m", "1Mi", "40m", "4Mi"),
			},
			e: makeRes("20m", "2Mi"),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c, m, g := cosRequests(u.cc)
			assert.True(t, c.Equal(*u.e.Cpu()))
			assert.True(t, m.Equal(*u.e.Memory()))
			assert.True(t, g.Equal(*extractGPU(u.e)))
		})
	}
}

// Helpers...

func makeContainer(n string, restartable bool, rc, rm, lc, lm string) v1.Container {
	always := v1.ContainerRestartPolicyAlways
	rq := v1.ResourceRequirements{
		Requests: makeRes(rc, rm),
		Limits:   makeRes(lc, lm),
	}
	var rp *v1.ContainerRestartPolicy
	if restartable {
		rp = &always
	}

	return v1.Container{Name: n, Resources: rq, RestartPolicy: rp}
}

func makeRes(c, m string) v1.ResourceList {
	cpu, _ := res.ParseQuantity(c)
	mem, _ := res.ParseQuantity(m)
	gpu, _ := res.ParseQuantity(c)

	return v1.ResourceList{
		v1.ResourceCPU:                    cpu,
		v1.ResourceMemory:                 mem,
		v1.ResourceName("nvidia.com/gpu"): gpu,
	}
}

func makeCoMX(n, c, m string) mv1beta1.ContainerMetrics {
	return mv1beta1.ContainerMetrics{
		Name:  n,
		Usage: makeRes(c, m),
	}
}

func testTime() time.Time {
	t, err := time.Parse(time.RFC3339, "2018-12-14T10:36:43.326972-07:00")
	if err != nil {
		fmt.Println("TestTime Failed", err)
	}
	return t
}
