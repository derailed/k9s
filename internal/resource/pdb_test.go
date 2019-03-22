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

func TestPDBListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewPDBList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, ns, l.GetNamespace())
	assert.Equal(t, "pdb", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestPDBHeader(t *testing.T) {
	row := resource.Row{
		"NAME",
		"MIN AVAILABLE",
		"MAX_ UNAVAILABLE",
		"ALLOWED DISRUPTIONS",
		"CURRENT",
		"DESIRED",
		"EXPECTED",
		"AGE",
	}
	assert.Equal(t, row, newPDB().Header(resource.DefaultNamespace))
}

func TestPDBFields(t *testing.T) {
	r := newPDB().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestPDBMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sPDB(), nil)

	cm := resource.NewPDBWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, pdbYaml(), ma)
}

func TestPDBListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sPDB()}, nil)

	l := resource.NewPDBListWithArgs("-", resource.NewPDBWithArgs(ca))
	// Make sure we can get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile()
		assert.Nil(t, err)
	}

	ca.VerifyWasCalled(m.Times(2)).List(resource.NotNamespaced)
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	assert.False(t, l.HasXRay())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 8, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

func TestPDBListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sPDB(), nil)
	l := resource.NewPDBListWithArgs("blee", resource.NewPDBWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
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
	return resource.NewPDB().NewInstance(k8sPDB())
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
