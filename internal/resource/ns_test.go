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

func TestNamespaceListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewNamespaceList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	assert.Equal(t, "ns", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestNamespaceHeader(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "STATUS", "AGE"}, newNamespace().Header(resource.DefaultNamespace))
}

func TestNamespaceFields(t *testing.T) {
	r := newNamespace().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestNamespaceMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("", "fred")).ThenReturn(k8sNamespace(), nil)

	cm := resource.NewNamespaceWithArgs(ca)
	ma, err := cm.Marshal("fred")
	ca.VerifyWasCalledOnce().Get("", "fred")
	assert.Nil(t, err)
	assert.Equal(t, nsYaml(), ma)
}

func TestNamespaceListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sNamespace()}, nil)

	l := resource.NewNamespaceListWithArgs("-", resource.NewNamespaceWithArgs(ca))
	// Make sure we can get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile()
		assert.Nil(t, err)
	}

	ca.VerifyWasCalled(m.Times(2)).List(resource.NotNamespaced)
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 3, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// Helpers...

func k8sNamespace() *v1.Namespace {
	return &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
	}
}

func newNamespace() resource.Columnar {
	return resource.NewNamespace().NewInstance(k8sNamespace())
}

func nsYaml() string {
	return `apiVersion: v1
kind: Namespace
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
spec: {}
status: {}
`
}
