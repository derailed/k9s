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

func NewServiceAccountListWithArgs(ns string, r *resource.ServiceAccount) resource.List {
	return resource.NewList(ns, "sa", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewServiceAccountWithArgs(conn k8s.Connection, res resource.Cruder) *resource.ServiceAccount {
	r := &resource.ServiceAccount{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestSaListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	l := NewServiceAccountListWithArgs(resource.AllNamespaces, NewServiceAccountWithArgs(mc, mr))
	l.SetNamespace(ns)

	assert.Equal(t, ns, l.GetNamespace())
	assert.Equal(t, "sa", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestSaHeader(t *testing.T) {
	s := newSa()
	e := append(resource.Row{"NAMESPACE"}, saHeader()...)
	assert.Equal(t, e, s.Header(resource.AllNamespaces))
	assert.Equal(t, saHeader(), s.Header("fred"))
}

func TestSaFields(t *testing.T) {
	uu := []struct {
		i resource.Columnar
		e resource.Row
	}{
		{i: newSa(), e: resource.Row{"blee", "fred", "1"}},
	}

	for _, u := range uu {
		assert.Equal(t, "blee/fred", u.i.Name())
		assert.Equal(t, u.e, u.i.Fields(resource.AllNamespaces)[:3])
		assert.Equal(t, u.e[1:], u.i.Fields("blee")[:2])
	}
}

func TestSAMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sSA(), nil)

	cm := NewServiceAccountWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")
	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, saYaml(), ma)
}

func TestSAListData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List("blee", metav1.ListOptions{})).ThenReturn(k8s.Collection{*k8sSA()}, nil)

	l := NewServiceAccountListWithArgs("blee", NewServiceAccountWithArgs(mc, mr))
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
	assert.Equal(t, 3, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// Helpers...

func k8sSA() *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "fred",
			Namespace:         "blee",
			CreationTimestamp: metav1.Time{testTime()},
		},
		Secrets: []v1.ObjectReference{{Name: "blee"}},
	}
}

func newSa() resource.Columnar {
	mc := NewMockConnection()
	return resource.NewServiceAccount(mc).New(k8sSA())
}

func saHeader() resource.Row {
	return resource.Row{"NAME", "SECRET", "AGE"}
}

func saYaml() string {
	return `apiVersion: v1
kind: ServiceAccount
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
secrets:
- name: blee
`
}
