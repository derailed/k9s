package resource_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewClusterRoleListWithArgs(ns string, r *resource.ClusterRole) resource.List {
	return resource.NewList(resource.NotNamespaced, "clusterrole", r, resource.CRUDAccess|resource.DescribeAccess)
}

func NewClusterRoleWithArgs(mc resource.Connection, res resource.Cruder) *resource.ClusterRole {
	r := &resource.ClusterRole{Base: resource.NewBase(mc, res)}
	r.Factory = r
	return r
}

func TestCRListAccess(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	ns := "blee"
	r := NewClusterRoleWithArgs(mc, mr)
	l := NewClusterRoleListWithArgs(resource.AllNamespaces, r)
	l.SetNamespace(ns)

	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	assert.Equal(t, "clusterrole", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestCRFields(t *testing.T) {
	r := newClusterRole().Fields("blee")
	assert.Equal(t, "fred", r[0])
}

func TestCRFieldsAllNS(t *testing.T) {
	r := newClusterRole().Fields(resource.AllNamespaces)
	assert.Equal(t, "fred", r[0])
}

func TestCRMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sCR(), nil)

	cr := NewClusterRoleWithArgs(mc, mr)
	ma, err := cr.Marshal("blee/fred")

	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, mrYaml(), ma)
}

func TestCRListData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List(resource.NotNamespaced, metav1.ListOptions{})).ThenReturn(k8s.Collection{*k8sCR()}, nil)

	l := NewClusterRoleListWithArgs("-", NewClusterRoleWithArgs(mc, mr))
	// Make sure we mcn get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile(nil, nil)
		assert.Nil(t, err)
	}

	mr.VerifyWasCalled(m.Times(2)).List(resource.NotNamespaced, metav1.ListOptions{})

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

func k8sCR() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "fred",
			Namespace:         "blee",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Rules: []rbacv1.PolicyRule{
			{
				Verbs:         []string{"get", "list"},
				APIGroups:     []string{""},
				ResourceNames: []string{"pod"},
			},
		},
	}
}

func newClusterRole() resource.Columnar {
	conn := NewMockConnection()
	return resource.NewClusterRole(conn).New(k8sCR())
}

func testTime() time.Time {
	t, err := time.Parse(time.RFC3339, "2018-12-14T10:36:43.326972-07:00")
	if err != nil {
		fmt.Println("TestTime Failed", err)
	}
	return t
}

func mrYaml() string {
	return `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
rules:
- apiGroups:
  - ""
  resourceNames:
  - pod
  verbs:
  - get
  - list
`
}
