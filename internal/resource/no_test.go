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

func TestNodeListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewNodeList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	assert.Equal(t, "no", l.GetName())
	for _, a := range []int{resource.ViewAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestNodeFields(t *testing.T) {
	r := newNode().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestNodeMarshal(t *testing.T) {
	setup(t)

	mx := NewMockMetricsIfc()
	m.When(mx.PerNodeMetrics([]v1.Node{*k8sNode()})).
		ThenReturn(map[string]k8s.Metric{"fred": {}}, nil)
	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sNode(), nil)

	cm := resource.NewNodeWithArgs(ca, mx)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, noYaml(), ma)
}

func TestNodeListData(t *testing.T) {
	setup(t)

	mx := NewMockMetricsIfc()
	m.When(mx.PerNodeMetrics([]v1.Node{*k8sNode()})).
		ThenReturn(map[string]k8s.Metric{"fred": {}}, nil)
	ca := NewMockCaller()
	m.When(ca.List("")).ThenReturn(k8s.Collection{*k8sNode()}, nil)

	l := resource.NewNodeListWithArgs("", resource.NewNodeWithArgs(ca, mx))
	// Make sure we can get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile()
		assert.Nil(t, err)
	}

	ca.VerifyWasCalled(m.Times(2)).List("")
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	assert.False(t, l.HasXRay())
	row := td.Rows["fred"]
	assert.Equal(t, 11, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

func TestNodeListDescribe(t *testing.T) {
	setup(t)

	mx := NewMockMetricsIfc()
	m.When(mx.PerNodeMetrics([]v1.Node{*k8sNode()})).
		ThenReturn(map[string]k8s.Metric{"fred": {}}, nil)
	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sNode(), nil)
	l := resource.NewNodeListWithArgs("blee", resource.NewNodeWithArgs(ca, mx))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
}

// Helpers...

func k8sNode() *v1.Node {
	return &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1.NodeSpec{},
		Status: v1.NodeStatus{
			Addresses: []v1.NodeAddress{
				{Address: "1.1.1.1"},
			},
		},
	}
}

func newNode() resource.Columnar {
	return resource.NewNode().NewInstance(k8sNode())
}

func noYaml() string {
	return `apiVersion: v1
kind: Node
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
spec: {}
status:
  addresses:
  - address: 1.1.1.1
    type: ""
  daemonEndpoints:
    kubeletEndpoint:
      Port: 0
  nodeInfo:
    architecture: ""
    bootID: ""
    containerRuntimeVersion: ""
    kernelVersion: ""
    kubeProxyVersion: ""
    kubeletVersion: ""
    machineID: ""
    operatingSystem: ""
    osImage: ""
    systemUUID: ""
`
}
