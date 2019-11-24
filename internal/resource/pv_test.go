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

func NewPVListWithArgs(ns string, r *resource.PersistentVolume) resource.List {
	return resource.NewList(resource.NotNamespaced, "pv", r, resource.CRUDAccess|resource.DescribeAccess)
}

func NewPVWithArgs(conn k8s.Connection, res resource.Cruder) *resource.PersistentVolume {
	r := &resource.PersistentVolume{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestPVListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	l := NewPVListWithArgs(resource.AllNamespaces, NewPVWithArgs(mc, mr))
	l.SetNamespace(ns)

	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	assert.Equal(t, "pv", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestPVFields(t *testing.T) {
	r := newPV().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestPVMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sPV(), nil)

	cm := NewPVWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")
	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, pvYaml(), ma)
}

func TestPVListData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List(resource.NotNamespaced, metav1.ListOptions{})).ThenReturn(k8s.Collection{*k8sPV()}, nil)

	l := NewPVListWithArgs("-", NewPVWithArgs(mc, mr))
	// Make sure we mrn get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile(nil, nil)
		assert.Nil(t, err)
	}

	mr.VerifyWasCalled(m.Times(2)).List(resource.NotNamespaced, metav1.ListOptions{})
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 9, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// Helpers...

func k8sPV() *v1.PersistentVolume {
	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1.PersistentVolumeSpec{},
	}
}

func newPV() resource.Columnar {
	mc := NewMockConnection()
	c, _ := resource.NewPersistentVolume(mc).New(k8sPV())
	return c
}

func pvYaml() string {
	return `apiVersion: v1
kind: PersistentVolume
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
spec: {}
status: {}
`
}
