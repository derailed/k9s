package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestPodListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewPodList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "po", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestPodHeader(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "READY", "STATUS", "RESTARTS", "CPU", "MEM", "IP", "NODE", "QOS", "AGE"}, newPod().Header(resource.DefaultNamespace))
}

func TestPodFields(t *testing.T) {
	r := newPod().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestPodMarshal(t *testing.T) {
	setup(t)

	mx := NewMockMetricsIfc()
	m.When(mx.PodMetrics()).ThenReturn(map[string]k8s.Metric{"fred": {}}, nil)
	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sPod(), nil)

	cm := resource.NewPodWithArgs(ca, mx)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, poYaml(), ma)
}

func TestPodListData(t *testing.T) {
	setup(t)

	mx := NewMockMetricsIfc()
	m.When(mx.PodMetrics()).ThenReturn(map[string]k8s.Metric{"fred": {}}, nil)
	ca := NewMockCaller()
	m.When(ca.List("")).ThenReturn(k8s.Collection{*k8sPod()}, nil)

	l := resource.NewPodListWithArgs("", resource.NewPodWithArgs(ca, mx))
	// Make sure we can get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile()
		assert.Nil(t, err)
	}

	ca.VerifyWasCalled(m.Times(2)).List(resource.AllNamespaces)
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, resource.AllNamespaces, l.GetNamespace())
	assert.True(t, l.HasXRay())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 11, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"blee"}, row.Fields[:1])
}

func TestPodListDescribe(t *testing.T) {
	setup(t)

	mx := NewMockMetricsIfc()
	m.When(mx.PodMetrics()).ThenReturn(map[string]k8s.Metric{"fred": {}}, nil)
	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sPod(), nil)
	l := resource.NewPodListWithArgs("blee", resource.NewPodWithArgs(ca, mx))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 8, len(props))
}

// Helpers...

func k8sPod() *v1.Pod {
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
	return resource.NewPod().NewInstance(k8sPod())
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
