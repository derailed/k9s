package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewStatefulSetListWithArgs(ns string, r *resource.StatefulSet) resource.List {
	return resource.NewList(ns, "sts", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewStatefulSetWithArgs(conn k8s.Connection, res resource.Cruder) *resource.StatefulSet {
	r := &resource.StatefulSet{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestStsListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	l := NewStatefulSetListWithArgs(resource.AllNamespaces, NewStatefulSetWithArgs(mc, mr))
	l.SetNamespace(ns)

	assert.Equal(t, l.GetNamespace(), ns)
	assert.Equal(t, "sts", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestStsHeader(t *testing.T) {
	s := newSts()
	e := append(resource.Row{"NAMESPACE"}, stsHeader()...)
	assert.Equal(t, e, s.Header(resource.AllNamespaces))
	assert.Equal(t, stsHeader(), s.Header("fred"))
}

func TestStsFields(t *testing.T) {
	uu := []struct {
		i resource.Columnar
		e resource.Row
	}{
		{i: newSts(), e: resource.Row{"blee", "fred", "0", "1"}},
	}

	for _, u := range uu {
		assert.Equal(t, "blee/fred", u.i.Name())
		assert.Equal(t, u.e, u.i.Fields(resource.AllNamespaces)[:4])
		assert.Equal(t, u.e[1:4], u.i.Fields("blee")[:3])
	}
}

func TestSTSMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sSTS(), nil)

	cm := NewStatefulSetWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")

	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, stsYaml(), ma)
}

func TestSTSListData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List("blee", metav1.ListOptions{})).ThenReturn(k8s.Collection{*k8sSTS()}, nil)

	l := NewStatefulSetListWithArgs("blee", NewStatefulSetWithArgs(mc, mr))
	// Make sure we mrn get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile(nil, nil)
		assert.Nil(t, err)
	}

	mr.VerifyWasCalled(m.Times(2)).List("blee", metav1.ListOptions{})
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, "blee", l.GetNamespace())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 4, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// Helpers...

func k8sSTS() *v1.StatefulSet {
	return &v1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "fred",
			Namespace:         "blee",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1.StatefulSetSpec{
			Replicas: new(int32),
		},
		Status: v1.StatefulSetStatus{
			ReadyReplicas: 1,
		},
	}
}

func newSts() resource.Columnar {
	mc := NewMockConnection()
	c, _ := resource.NewStatefulSet(mc).New(k8sSTS())
	return c
}

func stsHeader() resource.Row {
	return resource.Row{"NAME", "DESIRED", "CURRENT", "AGE"}
}

func stsYaml() string {
	return `apiVersion: apps/v1
kind: StatefulSet
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
spec:
  replicas: 0
  selector: null
  serviceName: ""
  template:
    metadata:
      creationTimestamp: null
    spec:
      containers: null
  updateStrategy: {}
status:
  readyReplicas: 1
  replicas: 0
`
}
