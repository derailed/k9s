// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	res "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func init() {
	render.AddColor = tcell.ColorBlue
	render.HighlightColor = tcell.ColorYellow
	render.CompletedColor = tcell.ColorGray
	render.StdColor = tcell.ColorWhite
	render.ErrColor = tcell.ColorRed
	render.KillColor = tcell.ColorGray
}

func TestPodColorer(t *testing.T) {
	stdHeader := render.Header{
		render.HeaderColumn{Name: "NAMESPACE"},
		render.HeaderColumn{Name: "NAME"},
		render.HeaderColumn{Name: "READY"},
		render.HeaderColumn{Name: "RESTARTS"},
		render.HeaderColumn{Name: "STATUS"},
		render.HeaderColumn{Name: "VALID"},
	}

	uu := map[string]struct {
		re render.RowEvent
		h  render.Header
		e  tcell.Color
	}{
		"valid": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", render.Running, ""},
				},
			},
			e: render.StdColor,
		},
		"init": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", render.PodInitializing, ""},
				},
			},
			e: render.AddColor,
		},
		"init-err": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", render.PodInitializing, "blah"},
				},
			},
			e: render.AddColor,
		},
		"initialized": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", render.Initialized, "blah"},
				},
			},
			e: render.HighlightColor,
		},
		"completed": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", render.Completed, "blah"},
				},
			},
			e: render.CompletedColor,
		},
		"terminating": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", render.Terminating, "blah"},
				},
			},
			e: render.KillColor,
		},
		"invalid": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", "Running", "blah"},
				},
			},
			e: render.ErrColor,
		},
		"unknown-cool": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", "blee", ""},
				},
			},
			e: render.AddColor,
		},
		"unknown-err": {
			h: stdHeader,
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", "blee", "doh"},
				},
			},
			e: render.ErrColor,
		},
		"status": {
			h: stdHeader[0:3],
			re: render.RowEvent{
				Kind: render.EventDelete,
				Row: render.Row{
					Fields: render.Fields{"blee", "fred", "1/1", "0", "blee", ""},
				},
			},
			e: render.KillColor,
		},
	}

	var r render.Pod
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, r.ColorerFunc()("", u.h, u.re))
		})
	}
}

func TestPodRender(t *testing.T) {
	pom := render.PodWithMetrics{
		Raw: load(t, "po"),
		MX:  makePodMX("nginx", "100m", "50Mi"),
	}

	var po render.Pod
	r := render.NewRow(14)
	err := po.Render(&pom, "", &r)
	assert.Nil(t, err)

	assert.Equal(t, "default/nginx", r.ID)
	e := render.Fields{"default", "nginx", "●", "1/1", "Running", "0", "172.17.0.6", "minikube", "<none>", "<none>", "100", "50", "100:0", "70:170", "100", "n/a", "71"}
	assert.Equal(t, e, r.Fields[:17])
}

func BenchmarkPodRender(b *testing.B) {
	pom := render.PodWithMetrics{
		Raw: load(b, "po"),
		MX:  makePodMX("nginx", "10m", "10Mi"),
	}
	var po render.Pod
	r := render.NewRow(12)

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
	r := render.NewRow(14)
	err := po.Render(&pom, "", &r)
	assert.Nil(t, err)

	assert.Equal(t, "default/nginx", r.ID)
	e := render.Fields{"default", "nginx", "●", "1/1", "Init:0/1", "0", "172.17.0.6", "minikube", "<none>", "<none>", "10", "10", "100:0", "70:170", "10", "n/a", "14"}
	assert.Equal(t, e, r.Fields[:17])
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
		"backoff": {
			pod: v1.Pod{
				Status: v1.PodStatus{
					Phase:                 v1.PodRunning,
					InitContainerStatuses: []v1.ContainerStatus{},
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
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, render.PodStatus(&u.pod))
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

// apiVersion: v1
// kind: Pod
// metadata:
//   creationTimestamp: "2023-11-11T17:01:40Z"
//   finalizers:
//   - batch.kubernetes.io/job-tracking
//   generateName: hello-28328646-
//   labels:
//     batch.kubernetes.io/controller-uid: 35cf5552-7180-48c1-b7b2-8b6e630a7860
//     batch.kubernetes.io/job-name: hello-28328646
//     controller-uid: 35cf5552-7180-48c1-b7b2-8b6e630a7860
//     job-name: hello-28328646
//   name: hello-28328646-h9fnh
//   namespace: fred
//   ownerReferences:
//   - apiVersion: batch/v1
//     blockOwnerDeletion: true
//     controller: true
//     kind: Job
//     name: hello-28328646
//     uid: 35cf5552-7180-48c1-b7b2-8b6e630a7860
//   resourceVersion: "381637"
//   uid: ea77c360-6375-459b-8b30-2ac0c59404cd
// spec:
//   containers:
//   - args:
//     - /bin/bash
//     - -c
//     - for i in {1..5}; do echo "hello";sleep 1; done
//     image: blang/busybox-bash
//     imagePullPolicy: Always
//     name: c1
//     resources: {}
//     terminationMessagePath: /dev/termination-log
//     terminationMessagePolicy: File
//     volumeMounts:
//     - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
//       name: kube-api-access-7sztm
//       readOnly: true
//   dnsPolicy: ClusterFirst
//   enableServiceLinks: true
//   nodeName: kind-worker
//   preemptionPolicy: PreemptLowerPriority
//   priority: 0
//   restartPolicy: OnFailure
//   schedulerName: default-scheduler
//   securityContext: {}
//   serviceAccount: default
//   serviceAccountName: default
//   terminationGracePeriodSeconds: 30
//   tolerations:
//   - effect: NoExecute
//     key: node.kubernetes.io/not-ready
//     operator: Exists
//     tolerationSeconds: 300
//   - effect: NoExecute
//     key: node.kubernetes.io/unreachable
//     operator: Exists
//     tolerationSeconds: 300
//   volumes:
//   - name: kube-api-access-7sztm
//     projected:
//       defaultMode: 420
//       sources:
//       - serviceAccountToken:
//           expirationSeconds: 3607
//           path: token
//       - configMap:
//           items:
//           - key: ca.crt
//             path: ca.crt
//           name: kube-root-ca.crt
//       - downwardAPI:
//           items:
//           - fieldRef:
//               apiVersion: v1
//               fieldPath: metadata.namespace
//             path: namespace
// status:
//   conditions:
//   - lastProbeTime: null
//     lastTransitionTime: "2023-11-11T17:01:40Z"
//     status: "True"
//     type: Initialized
//   - lastProbeTime: null
//     lastTransitionTime: "2023-11-11T17:01:40Z"
//     message: 'containers with unready status: [c1[]'
//     reason: ContainersNotReady
//     status: "False"
//     type: Ready
//   - lastProbeTime: null
//     lastTransitionTime: "2023-11-11T17:01:40Z"
//     message: 'containers with unready status: [c1[]'
//     reason: ContainersNotReady
//     status: "False"
//     type: ContainersReady
//   - lastProbeTime: null
//     lastTransitionTime: "2023-11-11T17:01:40Z"
//     status: "True"
//     type: PodScheduled
//   containerStatuses:
//   - image: blang/busybox-bash
//     imageID: ""
//     lastState: {}
//     name: c1
//     ready: false
//     restartCount: 0
//     started: false
//     state:
//       waiting:
//         message: Back-off pulling image "blang/busybox-bash"
//         reason: ImagePullBackOff
//   hostIP: 172.18.0.3
//   phase: Pending
//   podIP: 10.244.1.59
//   podIPs:
//   - ip: 10.244.1.59
//   qosClass: BestEffort
//   startTime: "2023-11-11T17:01:40Z"
