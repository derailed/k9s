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

func NewRBListWithArgs(ns string, r *resource.RoleBinding) resource.List {
	return resource.NewList(ns, "rb", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewRBWithArgs(conn k8s.Connection, res resource.Cruder) *resource.RoleBinding {
	r := &resource.RoleBinding{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

func TestRBMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sRB(), nil)

	cm := NewRBWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")

	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, rbYaml(), ma)
}

func TestRBListData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List("blee")).ThenReturn(k8s.Collection{*k8sRB()}, nil)

	l := NewRBListWithArgs("blee", NewRBWithArgs(mc, mr))
	// Make sure we mrn get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile(nil, nil)
		assert.Nil(t, err)
	}

	mr.VerifyWasCalled(m.Times(2)).List("blee")
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, "blee", l.GetNamespace())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 5, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

// Helpers...

func k8sRB() *v1.RoleBinding {
	return &v1.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Subjects: []v1.Subject{
			{
				Kind:      v1.UserKind,
				Name:      "fred",
				Namespace: "blee",
			},
		},
		RoleRef: v1.RoleRef{
			Kind: v1.UserKind,
			Name: "duh",
		},
	}
}

func newRB() resource.Columnar {
	mc := NewMockConnection()
	return resource.NewRoleBinding(mc).New(k8sRB())
}

func rbYaml() string {
	return `apiVersion: rbac.authorization.k8s.io/v1
kind: RoleBinding
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
roleRef:
  apiGroup: ""
  kind: User
  name: duh
subjects:
- kind: User
  name: fred
  namespace: blee
`
}
