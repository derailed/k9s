package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	res "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func NewNodeListWithArgs(ns string, r *resource.Node) resource.List {
	return resource.NewList(resource.NotNamespaced, "no", r, resource.ViewAccess|resource.DescribeAccess)
}

func NewNodeWithArgs(conn k8s.Connection, res resource.Cruder, mx resource.MetricsServer) *resource.Node {
	r := &resource.Node{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestNodeListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	mx := NewMockMetricsServer()

	ns := "blee"
	l := NewNodeListWithArgs(resource.AllNamespaces, NewNodeWithArgs(mc, mr, mx))
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
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sNode(), nil)
	mx := NewMockMetricsServer()

	cm := NewNodeWithArgs(mc, mr, mx)
	ma, err := cm.Marshal("blee/fred")

	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, noYaml(), ma)
}

func TestNodeListData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List("-", metav1.ListOptions{})).ThenReturn(k8s.Collection{*k8sNode()}, nil)
	mx := NewMockMetricsServer()
	m.When(mx.HasMetrics()).ThenReturn(true)
	m.When(mx.FetchNodesMetrics()).
		ThenReturn(&mv1beta1.NodeMetricsList{Items: []mv1beta1.NodeMetrics{makeMxNode("fred", "100m", "100Mi")}}, nil)

	l := NewNodeListWithArgs("-", NewNodeWithArgs(mc, mr, mx))
	// Make sure we mrn get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile(nil, nil)
		assert.Nil(t, err)
	}

	mr.VerifyWasCalled(m.Times(2)).List("-", metav1.ListOptions{})
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	row, ok := td.Rows["fred"]
	assert.True(t, ok)
	assert.Equal(t, 14, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// ----------------------------------------------------------------------------
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

func makeMxNode(name, cpu, mem string) mv1beta1.NodeMetrics {
	return v1beta1.NodeMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Usage: makeRes(cpu, mem),
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

func newNode() resource.Columnar {
	mc := NewMockConnection()
	return resource.NewNode(mc).New(k8sNode())
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
