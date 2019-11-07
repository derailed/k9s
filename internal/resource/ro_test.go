package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewRoleListWithArgs(ns string, r *resource.Role) resource.List {
	return resource.NewList(ns, "ro", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewRoleWithArgs(conn k8s.Connection, res resource.Cruder) *resource.Role {
	r := &resource.Role{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestRoleMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sRole(), nil)

	cm := NewRoleWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")
	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, roleYaml(), ma)
}

func TestRoleListData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List("blee", metav1.ListOptions{})).ThenReturn(k8s.Collection{*k8sRole()}, nil)

	l := NewRoleListWithArgs("blee", NewRoleWithArgs(mc, mr))
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
	assert.Equal(t, 2, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// Helpers...

func k8sRole() *v1.Role {
	return &v1.Role{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
	}
}

func newRole() resource.Columnar {
	mc := NewMockConnection()
	return resource.NewRole(mc).New(k8sRole())
}

func roleYaml() string {
	return `apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
rules: null
`
}
