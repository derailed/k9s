// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	res "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func init() {
	model1.AddColor = tcell.ColorBlue
	model1.HighlightColor = tcell.ColorYellow
	model1.CompletedColor = tcell.ColorGray
	model1.StdColor = tcell.ColorWhite
	model1.ErrColor = tcell.ColorRed
	model1.KillColor = tcell.ColorGray
}

func TestPodColorer(t *testing.T) {
	stdHeader := model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "READY"},
		model1.HeaderColumn{Name: "RESTARTS"},
		model1.HeaderColumn{Name: "STATUS"},
		model1.HeaderColumn{Name: "VALID"},
	}

	uu := map[string]struct {
		re model1.RowEvent
		h  model1.Header
		e  tcell.Color
	}{
		"valid": {
			h: stdHeader,
			re: model1.RowEvent{
				Kind: model1.EventAdd,
				Row: model1.Row{
					Fields: model1.Fields{"blee", "fred", "1/1", "0", render.Running, ""},
				},
			},
			e: model1.StdColor,
		},
		"init": {
			h: stdHeader,
			re: model1.RowEvent{
				Kind: model1.EventAdd,
				Row: model1.Row{
					Fields: model1.Fields{"blee", "fred", "1/1", "0", render.PodInitializing, ""},
				},
			},
			e: model1.AddColor,
		},
		"init-err": {
			h: stdHeader,
			re: model1.RowEvent{
				Kind: model1.EventAdd,
				Row: model1.Row{
					Fields: model1.Fields{"blee", "fred", "1/1", "0", render.PodInitializing, "blah"},
				},
			},
			e: model1.AddColor,
		},
		"initialized": {
			h: stdHeader,
			re: model1.RowEvent{
				Kind: model1.EventAdd,
				Row: model1.Row{
					Fields: model1.Fields{"blee", "fred", "1/1", "0", render.Initialized, "blah"},
				},
			},
			e: model1.HighlightColor,
		},
		"completed": {
			h: stdHeader,
			re: model1.RowEvent{
				Kind: model1.EventAdd,
				Row: model1.Row{
					Fields: model1.Fields{"blee", "fred", "1/1", "0", render.Completed, "blah"},
				},
			},
			e: model1.CompletedColor,
		},
		"terminating": {
			h: stdHeader,
			re: model1.RowEvent{
				Kind: model1.EventAdd,
				Row: model1.Row{
					Fields: model1.Fields{"blee", "fred", "1/1", "0", render.Terminating, "blah"},
				},
			},
			e: model1.KillColor,
		},
		"invalid": {
			h: stdHeader,
			re: model1.RowEvent{
				Kind: model1.EventAdd,
				Row: model1.Row{
					Fields: model1.Fields{"blee", "fred", "1/1", "0", "Running", "blah"},
				},
			},
			e: model1.ErrColor,
		},
		"unknown-cool": {
			h: stdHeader,
			re: model1.RowEvent{
				Kind: model1.EventAdd,
				Row: model1.Row{
					Fields: model1.Fields{"blee", "fred", "1/1", "0", "blee", ""},
				},
			},
			e: model1.AddColor,
		},
		"unknown-err": {
			h: stdHeader,
			re: model1.RowEvent{
				Kind: model1.EventAdd,
				Row: model1.Row{
					Fields: model1.Fields{"blee", "fred", "1/1", "0", "blee", "doh"},
				},
			},
			e: model1.ErrColor,
		},
		"status": {
			h: stdHeader[0:3],
			re: model1.RowEvent{
				Kind: model1.EventDelete,
				Row: model1.Row{
					Fields: model1.Fields{"blee", "fred", "1/1", "0", "blee", ""},
				},
			},
			e: model1.KillColor,
		},
	}

	var r render.Pod
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, r.ColorerFunc()("", u.h, &u.re))
		})
	}
}

func TestPodRender(t *testing.T) {
	pom := render.PodWithMetrics{
		Raw: load(t, "po"),
		MX:  makePodMX("nginx", "100m", "50Mi"),
	}

	var po render.Pod
	r := model1.NewRow(14)
	err := po.Render(&pom, "", &r)
	assert.Nil(t, err)

	assert.Equal(t, "default/nginx", r.ID)
	e := model1.Fields{"default", "nginx", "0", "●", "1/1", "Running", "0", "<unknown>", "100", "50", "100:0", "70:170", "100", "n/a", "71", "29", "172.17.0.6", "minikube", "default", "<none>"}
	assert.Equal(t, e, r.Fields[:20])
}

func BenchmarkPodRender(b *testing.B) {
	pom := render.PodWithMetrics{
		Raw: load(b, "po"),
		MX:  makePodMX("nginx", "10m", "10Mi"),
	}
	var po render.Pod
	r := model1.NewRow(12)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = po.Render(&pom, "", &r)
	}
}

func TestPodInitRender(t *testing.T) {
	pom := render.PodWithMetrics{
		Raw: load(t, "po_init"),
		MX:  makePodMX("nginx", "10m", "10Mi"),
	}

	var po render.Pod
	r := model1.NewRow(14)
	err := po.Render(&pom, "", &r)
	assert.Nil(t, err)

	assert.Equal(t, "default/nginx", r.ID)
	e := model1.Fields{"default", "nginx", "0", "●", "1/1", "Init:0/1", "0", "<unknown>", "10", "10", "100:0", "70:170", "10", "n/a", "14", "5", "172.17.0.6", "minikube", "default", "<none>"}
	assert.Equal(t, e, r.Fields[:20])
}

func TestPodSidecarRender(t *testing.T) {
	pom := render.PodWithMetrics{
		Raw: load(t, "po_sidecar"),
		MX:  makePodMX("sleep", "100m", "40Mi"),
	}

	var po render.Pod
	r := model1.NewRow(14)
	err := po.Render(&pom, "", &r)
	assert.Nil(t, err)

	assert.Equal(t, "default/sleep", r.ID)
	e := model1.Fields{"default", "sleep", "0", "●", "1/1", "Running", "0", "<unknown>", "100", "40", "50:250", "50:80", "200", "40", "80", "50", "10.244.0.8", "kind-control-plane", "default", "<none>"}
	assert.Equal(t, e, r.Fields[:20])
}

func TestCheckPodStatus(t *testing.T) {
	uu := map[string]struct {
		pod v1.Pod
		e   string
	}{
		"unknown": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Phase: render.PhaseUnknown,
				},
			},
			e: render.PhaseUnknown,
		},
		"running": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Phase:                 v1.PodRunning,
					InitContainerStatuses: []v1.ContainerStatus{},
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name: "c1",
							State: v1.ContainerState{
								Running: &v1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			e: render.PhaseRunning,
		},
		"gated": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Conditions: []v1.PodCondition{
						{Type: v1.PodScheduled, Reason: v1.PodReasonSchedulingGated},
					},
					Phase:                 v1.PodRunning,
					InitContainerStatuses: []v1.ContainerStatus{},
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name: "c1",
							State: v1.ContainerState{
								Running: &v1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			e: v1.PodReasonSchedulingGated,
		},

		"backoff": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name: "c1",
							State: v1.ContainerState{
								Waiting: &v1.ContainerStateWaiting{
									Reason: render.PhaseImagePullBackOff,
								},
							},
						},
					},
				},
			},
			e: render.PhaseImagePullBackOff,
		},
		"backoff-init": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					InitContainerStatuses: []v1.ContainerStatus{
						{
							Name: "ic1",
							State: v1.ContainerState{
								Waiting: &v1.ContainerStateWaiting{
									Reason: render.PhaseImagePullBackOff,
								},
							},
						},
					},
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name: "c1",
							State: v1.ContainerState{
								Waiting: &v1.ContainerStateWaiting{
									Reason: render.PhaseImagePullBackOff,
								},
							},
						},
					},
				},
			},
			e: "Init:ImagePullBackOff",
		},

		"init-terminated-cool": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					InitContainerStatuses: []v1.ContainerStatus{
						{
							Name:  "ic1",
							State: v1.ContainerState{},
						},
					},
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name: "c1",
							State: v1.ContainerState{
								Waiting: &v1.ContainerStateWaiting{
									Reason: render.PhaseImagePullBackOff,
								},
							},
						},
					},
				},
			},
			e: "Init:0/0",
		},

		"init-terminated-reason": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					InitContainerStatuses: []v1.ContainerStatus{
						{
							Name: "ic1",
							State: v1.ContainerState{
								Terminated: &v1.ContainerStateTerminated{
									ExitCode: 1,
									Reason:   "blah",
								},
							},
						},
					},
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name: "c1",
							State: v1.ContainerState{
								Waiting: &v1.ContainerStateWaiting{
									Reason: render.PhaseImagePullBackOff,
								},
							},
						},
					},
				},
			},
			e: "Init:blah",
		},
		"init-terminated-sig": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					InitContainerStatuses: []v1.ContainerStatus{
						{
							Name: "ic1",
							State: v1.ContainerState{
								Terminated: &v1.ContainerStateTerminated{
									ExitCode: 2,
									Signal:   9,
								},
							},
						},
					},
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name: "c1",
							State: v1.ContainerState{
								Waiting: &v1.ContainerStateWaiting{
									Reason: render.PhaseImagePullBackOff,
								},
							},
						},
					},
				},
			},
			e: "Init:Signal:9",
		},
		"init-terminated-code": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					InitContainerStatuses: []v1.ContainerStatus{
						{
							Name: "ic1",
							State: v1.ContainerState{
								Terminated: &v1.ContainerStateTerminated{
									ExitCode: 2,
								},
							},
						},
					},
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name: "c1",
							State: v1.ContainerState{
								Waiting: &v1.ContainerStateWaiting{
									Reason: render.PhaseImagePullBackOff,
								},
							},
						},
					},
				},
			},
			e: "Init:ExitCode:2",
		},

		"co-reason": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name: "c1",
							State: v1.ContainerState{
								Terminated: &v1.ContainerStateTerminated{
									Reason: "blah",
								},
							},
						},
					},
				},
			},
			e: "blah",
		},
		"co-reason-ready": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
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
			},
			e: "Running",
		},
		"co-reason-completed": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Conditions: []v1.PodCondition{
						{Type: v1.PodReady, Status: v1.ConditionTrue},
					},
					Phase: render.PhaseCompleted,
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
			},
			e: "Running",
		},

		"co-sig": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name: "c1",
							State: v1.ContainerState{
								Terminated: &v1.ContainerStateTerminated{
									ExitCode: 2,
									Signal:   9,
								},
							},
						},
					},
				},
			},
			e: "Signal:9",
		},
		"co-code": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
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
			},
			e: "ExitCode:2",
		},
		"co-ready": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Phase: v1.PodRunning,
					ContainerStatuses: []v1.ContainerStatus{
						{
							Name: "c1",
							State: v1.ContainerState{
								Running: &v1.ContainerStateRunning{},
							},
						},
					},
				},
			},
			e: "Running",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, render.PodStatus(&u.pod))
		})
	}
}

func TestCheckPhase(t *testing.T) {
	always := v1.ContainerRestartPolicyAlways
	uu := map[string]struct {
		pod v1.Pod
		e   string
	}{
		"unknown": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Phase: render.PhaseUnknown,
				},
			},
			e: render.PhaseUnknown,
		},
		"terminating": {
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: testTime()},
				},
				Status: v1.PodStatus{
					Phase:  render.PhaseUnknown,
					Reason: "bla",
				},
			},
			e: render.PhaseTerminating,
		},
		"terminating-toast-node": {
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: testTime()},
				},
				Status: v1.PodStatus{
					Phase:  render.PhaseUnknown,
					Reason: render.NodeUnreachablePodReason,
				},
			},
			e: render.PhaseUnknown,
		},
		"restartable": {
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: testTime()},
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name:          "ic1",
							RestartPolicy: &always,
						},
					},
				},
				Status: v1.PodStatus{
					Phase:  render.PhaseUnknown,
					Reason: "bla",
					InitContainerStatuses: []v1.ContainerStatus{
						{
							Name: "ic1",
						},
					},
				},
			},
			e: "Init:0/1",
		},
		"waiting": {
			pod: v1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metav1.Time{Time: testTime()},
				},
				Spec: v1.PodSpec{
					InitContainers: []v1.Container{
						{
							Name:          "ic1",
							RestartPolicy: &always,
						},
					},
					Containers: []v1.Container{
						{
							Name: "c1",
						},
					},
				},
				Status: v1.PodStatus{
					Phase:  render.PhaseUnknown,
					Reason: "bla",
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
									Reason: "bla",
								},
							},
						},
					},
				},
			},
			e: "Init:0/1",
		},
	}

	var p render.Pod
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, p.Phase(&u.pod))
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makePodMX(name, cpu, mem string) *mv1beta1.PodMetrics {
	return &mv1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Containers: []mv1beta1.ContainerMetrics{
			{Usage: makeRes(cpu, mem)},
		},
	}
}

func makeRes(c, m string) v1.ResourceList {
	cpu, _ := res.ParseQuantity(c)
	mem, _ := res.ParseQuantity(m)

	return v1.ResourceList{
		v1.ResourceCPU:    cpu,
		v1.ResourceMemory: mem,
	}
}
