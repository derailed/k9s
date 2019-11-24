package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	resv1 "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewPVCListWithArgs(ns string, r *resource.PersistentVolumeClaim) resource.List {
	return resource.NewList(ns, "pvc", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewPVCWithArgs(conn k8s.Connection, res resource.Cruder) *resource.PersistentVolumeClaim {
	r := &resource.PersistentVolumeClaim{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestPVCListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	l := NewPVCListWithArgs(resource.AllNamespaces, NewPVCWithArgs(mc, mr))
	l.SetNamespace(ns)

	assert.Equal(t, "blee", l.GetNamespace())
	assert.Equal(t, "pvc", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestPVCFields(t *testing.T) {
	r := newPVC().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestPVCMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sPVC(), nil)

	cm := NewPVCWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")
	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, pvcYaml(), ma)
}

func TestPVCListData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List("blee", metav1.ListOptions{})).ThenReturn(k8s.Collection{*k8sPVC()}, nil)

	l := NewPVCListWithArgs("blee", NewPVCWithArgs(mc, mr))
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
	assert.Equal(t, 7, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// Helpers...

func k8sPVC() *v1.PersistentVolumeClaim {
	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1.PersistentVolumeClaimSpec{
			VolumeName: "duh",
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: resv1.Quantity{},
				},
			},
		},
	}
}

func newPVC() resource.Columnar {
	mc := NewMockConnection()
	c, _ := resource.NewPersistentVolumeClaim(mc).New(k8sPVC())
	return c
}

func pvcYaml() string {
	return `apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
spec:
  resources:
    requests:
      storage: "0"
  volumeName: duh
status: {}
`
}
