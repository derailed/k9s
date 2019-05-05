package resource_test

import (
	"testing"

	// 	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"

	// 	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewEndpointsListWithArgs(ns string, r *resource.Endpoints) resource.List {
	return resource.NewList(ns, "ep", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewEndpointsWithArgs(conn k8s.Connection, res resource.Cruder) *resource.Endpoints {
	r := &resource.Endpoints{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestEndpointsListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	l := NewEndpointsListWithArgs(resource.AllNamespaces, NewEndpointsWithArgs(mc, mr))
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "ep", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestEndpointsMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sEndpoints(), nil)

	cm := NewEndpointsWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")
	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, epYaml(), ma)
}

func TestEndpointsListData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sEndpoints()}, nil)

	l := NewEndpointsListWithArgs("-", NewEndpointsWithArgs(mc, mr))
	// Make sure we mrn get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile(nil, nil)
		assert.Nil(t, err)
	}

	mr.VerifyWasCalled(m.Times(2)).List(resource.NotNamespaced)
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 3, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// Helpers...

func k8sEndpoints() *v1.Endpoints {
	return &v1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Subsets: []v1.EndpointSubset{
			{
				Addresses: []v1.EndpointAddress{
					{IP: "1.1.1.1"},
				},
				Ports: []v1.EndpointPort{
					{Port: 80, Protocol: "TCP"},
				},
			},
		},
	}
}

func newEndpoints() resource.Columnar {
	mc := NewMockConnection()
	return resource.NewEndpoints(mc).New(k8sEndpoints())
}

func epYaml() string {
	return `apiVersion: v1
kind: Endpoint
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
subsets:
- addresses:
  - ip: 1.1.1.1
  ports:
  - port: 80
    protocol: TCP
`
}
