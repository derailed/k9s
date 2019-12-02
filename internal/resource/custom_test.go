package resource_test

import (
	"testing"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
)

func NewCustomListWithArgs(ns, name string, r *resource.Custom) resource.List {
	return resource.NewList(ns, name, r, resource.AllVerbsAccess)
}

func NewCustomWithArgs(conn k8s.Connection, res resource.Cruder) *resource.Custom {
	r := &resource.Custom{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestCustomListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	r := NewCustomWithArgs(mc, mr)
	l := NewCustomListWithArgs(resource.AllNamespaces, "fred", r)
	l.SetNamespace(ns)

	assert.Equal(t, ns, l.GetNamespace())
	assert.Equal(t, "fred", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestCustomFields(t *testing.T) {
	r := newCustom().Fields("blee")
	assert.Equal(t, "a", r[0])
}

// BOZO!!
// func TestCustomMarshal(t *testing.T) {
// 	mc := NewMockConnection()
// 	mr := NewMockCruder()
// 	m.When(mr.Get("blee", "fred")).ThenReturn(k8sCustomTable(), nil)

// 	cm := NewCustomWithArgs(mc, mr)
// 	ma, err := cm.Marshal("blee/fred")
// 	mr.VerifyWasCalledOnce().Get("blee", "fred")

// 	assert.Nil(t, err)
// 	assert.Equal(t, customYaml(), ma)
// }

func TestCustomMarshalWithUnstructured(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sUnstructured(), nil)

	cm := NewCustomWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")
	mr.VerifyWasCalledOnce().Get("blee", "fred")

	assert.Nil(t, err)
	assert.Equal(t, unstructuredYAML(), ma)
}

// BOZO!!
// func TestCustomListData(t *testing.T) {
// 	mc := NewMockConnection()
// 	mr := NewMockCruder()
// 	m.When(mr.List("blee", metav1.ListOptions{})).ThenReturn(k8s.Collection{k8sCustomTable()}, nil)

// 	l := NewCustomListWithArgs("blee", "fred", NewCustomWithArgs(mc, mr))
// 	// Make sure we can get deltas!
// 	for i := 0; i < 2; i++ {
// 		err := l.Reconcile(nil, "", "")
// 		assert.Nil(t, err)
// 	}

// 	mr.VerifyWasCalled(m.Times(2)).List("blee", metav1.ListOptions{})
// 	td := l.Data()
// 	assert.Equal(t, 1, len(td.Rows))
// 	assert.Equal(t, "blee", l.GetNamespace())
// 	row := td.Rows["blee/fred"]
// 	assert.Equal(t, 3, len(row.Deltas))
// 	for _, d := range row.Deltas {
// 		assert.Equal(t, "", d)
// 	}
// 	assert.Equal(t, resource.Row{"a"}, row.Fields[:1])
// }

// Helpers...

func k8sCustomTable() *metav1beta1.Table {
	return &metav1beta1.Table{
		ColumnDefinitions: []metav1beta1.TableColumnDefinition{
			{Name: "A"},
			{Name: "B"},
			{Name: "C"},
		},
		Rows: []metav1beta1.TableRow{
			{
				Object: runtime.RawExtension{
					Raw: []byte(`{
        "kind": "fred",
        "apiVersion": "v1",
        "metadata": {
          "namespace": "blee",
          "name": "fred"
        }}`),
				},
				Cells: []interface{}{
					"a",
					"b",
					"c",
				},
			},
		},
	}
}

func k8sUnstructured() *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"kind":       "fred",
			"apiVersion": "v1",
			"metadata": map[string]interface{}{
				"namespace": "blee",
				"name":      "fred",
			},
		},
	}
}

func unstructuredYAML() string {
	return `apiVersion: v1
kind: fred
metadata:
  name: fred
  namespace: blee
`
}

func k8sCustomRow() *metav1beta1.TableRow {
	return &metav1beta1.TableRow{
		Object: runtime.RawExtension{
			Raw: []byte(`{
        "kind": "fred",
        "apiVersion": "v1",
        "metadata": {
          "namespace": "blee",
          "name": "fred"
        }}`),
		},
		Cells: []interface{}{
			"a",
			"b",
			"c",
		},
	}
}

func newCustom() resource.Columnar {
	mc := NewMockConnection()
	c, _ := resource.NewCustom(mc, "g/v1/fred").New(k8sCustomRow())
	return c
}

func customYaml() string {
	return `typemeta:
  kind: ""
  apiversion: ""
listmeta:
  selflink: ""
  resourceversion: ""
  continue: ""
  remainingitemcount: null
columndefinitions:
- name: A
  type: ""
  format: ""
  description: ""
  priority: 0
- name: B
  type: ""
  format: ""
  description: ""
  priority: 0
- name: C
  type: ""
  format: ""
  description: ""
  priority: 0
rows:
- cells:
  - a
  - b
  - c
  conditions: []
  object:
    raw:
    - 123
    - 10
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 34
    - 107
    - 105
    - 110
    - 100
    - 34
    - 58
    - 32
    - 34
    - 102
    - 114
    - 101
    - 100
    - 34
    - 44
    - 10
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 34
    - 97
    - 112
    - 105
    - 86
    - 101
    - 114
    - 115
    - 105
    - 111
    - 110
    - 34
    - 58
    - 32
    - 34
    - 118
    - 49
    - 34
    - 44
    - 10
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 34
    - 109
    - 101
    - 116
    - 97
    - 100
    - 97
    - 116
    - 97
    - 34
    - 58
    - 32
    - 123
    - 10
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 34
    - 110
    - 97
    - 109
    - 101
    - 115
    - 112
    - 97
    - 99
    - 101
    - 34
    - 58
    - 32
    - 34
    - 98
    - 108
    - 101
    - 101
    - 34
    - 44
    - 10
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 34
    - 110
    - 97
    - 109
    - 101
    - 34
    - 58
    - 32
    - 34
    - 102
    - 114
    - 101
    - 100
    - 34
    - 10
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 32
    - 125
    - 125
    object: null
`
}
