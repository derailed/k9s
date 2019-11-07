package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func NewClusterRoleBindingListWithArgs(ns string, r *resource.ClusterRoleBinding) resource.List {
	return resource.NewList(resource.NotNamespaced, "clusterrolebinding", r, resource.ViewAccess|resource.DeleteAccess|resource.DescribeAccess)
}

func NewClusterRoleBindingWithArgs(conn k8s.Connection, res resource.Cruder) *resource.ClusterRoleBinding {
	r := &resource.ClusterRoleBinding{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestCRBFields(t *testing.T) {
	conn := NewMockConnection()

	r := newCRB(conn).Fields(resource.AllNamespaces)

	assert.Equal(t, "fred", r[0])
}

func TestCRBMarshal(t *testing.T) {
	conn := NewMockConnection()
	ca := NewMockCruder()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sCRB(), nil)

	cm := NewClusterRoleBindingWithArgs(conn, ca)
	ma, err := cm.Marshal("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, crbYaml(), ma)
}

func TestCRBListData(t *testing.T) {
	conn := NewMockConnection()
	ca := NewMockCruder()
	m.When(ca.List(resource.NotNamespaced, metav1.ListOptions{})).ThenReturn(k8s.Collection{*k8sCRB()}, nil)

	l := NewClusterRoleBindingListWithArgs("-", NewClusterRoleBindingWithArgs(conn, ca))
	// Make sure we can get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile(nil, nil)
		assert.Nil(t, err)
	}

	ca.VerifyWasCalled(m.Times(2)).List(resource.NotNamespaced, metav1.ListOptions{})
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	row := td.Rows["fred"]
	assert.Equal(t, 5, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// Helpers...

func k8sCRB() *rbacv1.ClusterRoleBinding {
	return &rbacv1.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "fred",
			Namespace:         "blee",
			CreationTimestamp: metav1.Time{testTime()},
		},
		Subjects: []rbacv1.Subject{
			{Kind: "test", Name: "fred", Namespace: "blee"},
		},
	}
}

func newCRB(c resource.Connection) resource.Columnar {
	return resource.NewClusterRoleBinding(c).New(k8sCRB())
}

func crbYaml() string {
	return `apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
roleRef:
  apiGroup: ""
  kind: ""
  name: ""
subjects:
- kind: test
  name: fred
  namespace: blee
`
}
