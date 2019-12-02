package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func NewPodListWithArgs(ns string, r *resource.Pod) resource.List {
	return resource.NewList(ns, "po", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewPodWithArgs(conn k8s.Connection, res resource.Cruder, mx resource.MetricsServer) *resource.Pod {
	r := &resource.Pod{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestPodListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	mx := NewMockMetricsServer()

	ns := "blee"
	l := NewPodListWithArgs(resource.AllNamespaces, NewPodWithArgs(mc, mr, mx))
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "po", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestPodFields(t *testing.T) {
	r := newPod().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestPodGatherMX(t *testing.T) {
	uu := map[string]struct {
		resources             v1.ResourceRequirements
		metrics               mv1beta1.PodMetrics
		expectedCpuPercentage string
		expectedMemPercentage string
	}{
		"request": {
			v1.ResourceRequirements{
				Requests: makeRes("500m", "512Mi"),
			},
			makeMxPod("p1", "250m", "256Mi"),
			"150",
			"150",
		},
		"limit": {
			v1.ResourceRequirements{
				Limits: makeRes("1000m", "1024Mi"),
			},
			makeMxPod("p2", "250m", "256Mi"),
			"75",
			"75",
		},
		"both": {
			v1.ResourceRequirements{
				Requests: makeRes("500m", "512Mi"),
				Limits:   makeRes("1000m", "1024Mi"),
			},
			makeMxPod("p3", "250m", "256Mi"),
			"150",
			"150",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			r := NewPodWithMetrics(u.metrics, u.resources).Fields("blee")

			assert.Equal(t, u.expectedCpuPercentage, r[6])
			assert.Equal(t, u.expectedMemPercentage, r[7])
		})
	}
}

// BOZO!!
// func TestPodMarshal(t *testing.T) {
// 	mc := NewMockConnection()
// 	mr := NewMockCruder()
// 	m.When(mr.Get("blee", "fred")).ThenReturn(makePod(), nil)
// 	mx := NewMockMetricsServer()

// 	cm := NewPodWithArgs(mc, mr, mx)
// 	ma, err := cm.Marshal("blee/fred")

// 	mr.VerifyWasCalledOnce().Get("blee", "fred")
// 	assert.Nil(t, err)
// 	assert.Equal(t, poYaml(), ma)
// }

// BOZO!!
// func TestPodListData(t *testing.T) {
// 	mc := NewMockConnection()
// 	mr := NewMockCruder()
// 	m.When(mr.List("blee", metav1.ListOptions{})).ThenReturn(k8s.Collection{*makePod()}, nil)
// 	mx := NewMockMetricsServer()
// 	m.When(mx.HasMetrics()).ThenReturn(true)
// 	m.When(mx.FetchPodsMetrics("blee")).
// 		ThenReturn(&mv1beta1.PodMetricsList{Items: []mv1beta1.PodMetrics{makeMxPod("p1", "100m", "20Mi")}}, nil)

// 	l := NewPodListWithArgs("blee", NewPodWithArgs(mc, mr, mx))
// 	// Make sure we mcn get deltas!
// 	for i := 0; i < 2; i++ {
// 		err := l.Reconcile(nil, "", "")
// 		assert.Nil(t, err)
// 	}

// 	mr.VerifyWasCalled(m.Times(2)).List("blee", metav1.ListOptions{})
// 	td := l.Data()
// 	assert.Equal(t, 1, len(td.Rows))
// 	assert.Equal(t, "blee", l.GetNamespace())
// 	row := td.Rows["blee/fred"]
// 	assert.Equal(t, 12, len(row.Deltas))
// 	for _, d := range row.Deltas {
// 		assert.Equal(t, "", d)
// 	}
// 	assert.Equal(t, "fred", strings.TrimSpace(row.Fields[:1][0]))
// }

func BenchmarkPodFields(b *testing.B) {
	p := resource.NewPod(nil)
	po := makePod()

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		pod, _ := p.New(po)
		pod.Fields("")
	}
}

// ----------------------------------------------------------------------------
// Helpers...
func makePodWithContainerSpec(resources v1.ResourceRequirements) *v1.Pod {
	pod := makePod()
	pod.Spec.Containers[0].Resources = resources
	return pod
}

func makePod() *v1.Pod {
	var i int32 = 1
	var t = v1.HostPathDirectory
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			Labels:            map[string]string{"blee": "duh"},
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1.PodSpec{
			Priority:          &i,
			PriorityClassName: "bozo",
			Containers: []v1.Container{
				{
					Name:  "fred",
					Image: "blee",
					Env: []v1.EnvVar{
						{
							Name:  "fred",
							Value: "1",
							ValueFrom: &v1.EnvVarSource{
								ConfigMapKeyRef: &v1.ConfigMapKeySelector{Key: "blee"},
							},
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "fred",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/blee",
							Type: &t,
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: "Running",
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:         "fred",
					State:        v1.ContainerState{Running: &v1.ContainerStateRunning{}},
					RestartCount: 0,
				},
			},
		},
	}
}

func newPod() resource.Columnar {
	mc := NewMockConnection()
	c, _ := resource.NewPod(mc).New(makePod())
	return c
}

func NewPodWithMetrics(metrics mv1beta1.PodMetrics, resources v1.ResourceRequirements) resource.Columnar {
	mc := NewMockConnection()
	p := resource.NewPod(mc)
	r, _ := p.New(makePodWithContainerSpec(resources))
	r.SetPodMetrics(&metrics)
	return r
}

func poYaml() string {
	return `apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  labels:
    blee: duh
  name: fred
  namespace: blee
spec:
  containers:
  - env:
    - name: fred
      value: "1"
      valueFrom:
        configMapKeyRef:
          key: blee
    image: blee
    name: fred
    resources: {}
  priority: 1
  priorityClassName: bozo
  volumes:
  - hostPath:
      path: /blee
      type: Directory
    name: fred
status:
  containerStatuses:
  - image: ""
    imageID: ""
    lastState: {}
    name: fred
    ready: false
    restartCount: 0
    state:
      running:
        startedAt: null
  phase: Running
`
}

func makeMxPod(name, cpu, mem string) mv1beta1.PodMetrics {
	return mv1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Containers: []mv1beta1.ContainerMetrics{
			{Usage: makeRes(cpu, mem)},
			{Usage: makeRes(cpu, mem)},
			{Usage: makeRes(cpu, mem)},
		},
	}
}
