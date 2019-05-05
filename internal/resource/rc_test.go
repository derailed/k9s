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

func NewRCListWithArgs(ns string, r *resource.ReplicationController) resource.List {
	return resource.NewList(ns, "rc", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewRCWithArgs(conn k8s.Connection, res resource.Cruder) *resource.ReplicationController {
	r := &resource.ReplicationController{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestRCListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	l := NewRCListWithArgs(resource.AllNamespaces, NewRCWithArgs(mc, mr))
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "rc", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestRCFields(t *testing.T) {
	r := newRC().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestRCMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sRC(), nil)

	cm := NewRCWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")

	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, rcYaml(), ma)
}

func TestRCListData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List("blee")).ThenReturn(k8s.Collection{*k8sRC()}, nil)

	l := NewRCListWithArgs("blee", NewRCWithArgs(mc, mr))
	// Make sure we mrn get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile(nil, nil)
		assert.Nil(t, err)
	}

	mr.VerifyWasCalled(m.Times(2)).List("blee")
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, "blee", l.GetNamespace())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 5, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// Helpers...

func k8sRC() *v1.ReplicationController {
	var c int32 = 10
	return &v1.ReplicationController{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1.ReplicationControllerSpec{
			Replicas: &c,
		},
	}
}

func newRC() resource.Columnar {
	mc := NewMockConnection()
	return resource.NewReplicationController(mc).New(k8sRC())
}

func rcYaml() string {
	return `apiVersion: v1
kind: ReplicationController
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
spec:
  replicas: 10
status:
  replicas: 0
`
}
