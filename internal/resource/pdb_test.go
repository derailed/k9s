package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewPDBListWithArgs(ns string, r *resource.PodDisruptionBudget) resource.List {
	return resource.NewList(ns, "pdb", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewPDBWithArgs(conn k8s.Connection, res resource.Cruder) *resource.PodDisruptionBudget {
	r := &resource.PodDisruptionBudget{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestPDBListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	l := NewPDBListWithArgs(resource.AllNamespaces, NewPDBWithArgs(mc, mr))
	l.SetNamespace(ns)

	assert.Equal(t, ns, l.GetNamespace())
	assert.Equal(t, "pdb", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestPDBFields(t *testing.T) {
	r := newPDB().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestPDBMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sPDB(), nil)

	cm := NewPDBWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")
	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, pdbYaml(), ma)
}

func TestPDBListData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List("blee", metav1.ListOptions{})).ThenReturn(k8s.Collection{*k8sPDB()}, nil)

	l := NewPDBListWithArgs("blee", NewPDBWithArgs(mc, mr))
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
	assert.Equal(t, 8, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// Helpers...

func k8sPDB() *v1beta1.PodDisruptionBudget {
	return &v1beta1.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1beta1.PodDisruptionBudgetSpec{},
	}
}

func newPDB() resource.Columnar {
	mc := NewMockConnection()
	c, _ := resource.NewPDB(mc).New(k8sPDB())
	return c
}

func pdbYaml() string {
	return `apiVersion: v1beta1
kind: PodDisruptionBudget
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
spec: {}
status:
  currentHealthy: 0
  desiredHealthy: 0
  disruptionsAllowed: 0
  expectedPods: 0
`
}
