package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestCRDListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewCRDList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	assert.Equal(t, "crd", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestCRDHeader(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "AGE"}, newCRD().Header(resource.DefaultNamespace))
}

func TestCRDHeaderAllNS(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "AGE"}, newCRD().Header(resource.AllNamespaces))
}

func TestCRDFields(t *testing.T) {
	r := newCRD().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestCRDFieldsAllNS(t *testing.T) {
	r := newCRD().Fields(resource.AllNamespaces)
	assert.Equal(t, "fred", r[0])
}

func TestCRDMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sCRD(), nil)

	cm := resource.NewCRDWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, crdYaml(), ma)
}

func TestCRDListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sCRD()}, nil)

	l := resource.NewCRDListWithArgs("-", resource.NewCRDWithArgs(ca))
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
	row := td.Rows["fred"]
	assert.Equal(t, 2, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

func TestCRDListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sCRD(), nil)
	l := resource.NewCRDListWithArgs("blee", resource.NewCRDWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
}

// Helpers...

func k8sCRD() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"metadata": map[string]interface{}{
				"namespace":         "blee",
				"name":              "fred",
				"creationTimestamp": "2018-12-14T10:36:43.326972Z",
			},
		},
	}
}

func newCRD() resource.Columnar {
	return resource.NewCRD().NewInstance(k8sCRD())
}

func crdYaml() string {
	return `object:
  metadata:
    creationTimestamp: "2018-12-14T10:36:43.326972Z"
    name: fred
    namespace: blee
`
}
