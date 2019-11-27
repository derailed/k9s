package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewSCListWithArgs(ns string, r *resource.StorageClass) resource.List {
	return resource.NewList(resource.NotNamespaced, "sc", r, resource.CRUDAccess|resource.DescribeAccess)
}

func NewSCWithArgs(conn k8s.Connection, res resource.Cruder) *resource.StorageClass {
	r := &resource.StorageClass{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestSCListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	l := NewSCListWithArgs(resource.AllNamespaces, NewSCWithArgs(mc, mr))
	l.SetNamespace(ns)

	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	assert.Equal(t, "sc", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestSCFields(t *testing.T) {
	r := newSC().Fields("blee")
	assert.Equal(t, "storage-test", r[0])
}

func TestSCMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "storage-test")).ThenReturn(k8sSC(), nil)

	cm := NewSCWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/storage-test")
	mr.VerifyWasCalledOnce().Get("blee", "storage-test")
	assert.Nil(t, err)
	assert.Equal(t, scYaml(), ma)
}

// BOZO!!
// func TestSCListData(t *testing.T) {
// 	mc := NewMockConnection()
// 	mr := NewMockCruder()
// 	m.When(mr.List(resource.NotNamespaced, metav1.ListOptions{})).ThenReturn(k8s.Collection{*k8sSC()}, nil)

// 	l := NewSCListWithArgs("-", NewSCWithArgs(mc, mr))
// 	// Make sure we mrn get deltas!
// 	for i := 0; i < 2; i++ {
// 		err := l.Reconcile(nil, "", "")
// 		assert.Nil(t, err)
// 	}

// 	mr.VerifyWasCalled(m.Times(2)).List(resource.NotNamespaced, metav1.ListOptions{})
// 	td := l.Data()
// 	assert.Equal(t, 1, len(td.Rows))
// 	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
// 	row := td.Rows["storage-test"]
// 	assert.Equal(t, 3, len(row.Deltas))
// 	for _, d := range row.Deltas {
// 		assert.Equal(t, "", d)
// 	}
// 	assert.Equal(t, resource.Row{"storage-test"}, row.Fields[:1])
// }

// Helpers...

func k8sSC() *v1.StorageClass {
	return &v1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "storage-test",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
	}
}

func newSC() resource.Columnar {
	mc := NewMockConnection()
	c, _ := resource.NewStorageClass(mc).New(k8sSC())
	return c
}

func scYaml() string {
	return `apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: storage-test
provisioner: ""
`
}
