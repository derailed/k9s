package resource_test

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func NewCRDListWithArgs(ns string, r *resource.CustomResourceDefinition) resource.List {
	return resource.NewList("-", "crd", r, resource.CRUDAccess|resource.DescribeAccess)
}

func NewCRDWithArgs(conn k8s.Connection, res resource.Cruder) *resource.CustomResourceDefinition {
	r := &resource.CustomResourceDefinition{Base: resource.NewBase(conn, res)}
	r.Factory = r

	return r
}

func TestCRDListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	r := NewCRDWithArgs(mc, mr)
	l := NewCRDListWithArgs(resource.AllNamespaces, r)
	l.SetNamespace(ns)

	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	assert.Equal(t, "crd", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
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
	mc := NewMockConnection()
	cr := NewMockCruder()
	m.When(cr.Get("blee", "fred")).ThenReturn(k8sCRD(), nil)

	r := NewCRDWithArgs(mc, cr)
	ma, err := r.Marshal("blee/fred")

	cr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, crdYaml(), ma)
}

func TestCRDListData(t *testing.T) {
	mc := NewMockConnection()
	cr := NewMockCruder()
	m.When(cr.List(resource.NotNamespaced, v1.ListOptions{})).ThenReturn(k8s.Collection{*k8sCRD()}, nil)

	l := NewCRDListWithArgs("-", NewCRDWithArgs(mc, cr))
	// Make sure we can get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile(nil, nil)
		assert.Nil(t, err)
	}

	cr.VerifyWasCalled(m.Times(2)).List(resource.NotNamespaced, metav1.ListOptions{})
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	row := td.Rows["fred"]
	assert.Equal(t, 2, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
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
	mc := NewMockConnection()
	c, _ := resource.NewCustomResourceDefinition(mc).New(k8sCRD())
	return c
}

func crdYaml() string {
	return `object:
  metadata:
    creationTimestamp: "2018-12-14T10:36:43.326972Z"
    name: fred
    namespace: blee
`
}
