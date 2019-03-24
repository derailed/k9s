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

func TestRBMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sRB(), nil)

	cm := resource.NewRoleBindingWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, rbYaml(), ma)
}

func TestRBListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List("blee")).ThenReturn(k8s.Collection{*k8sRB()}, nil)

	l := resource.NewRoleBindingListWithArgs("blee", resource.NewRoleBindingWithArgs(ca))
	// Make sure we can get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile()
		assert.Nil(t, err)
	}

	ca.VerifyWasCalled(m.Times(2)).List("blee")
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, "blee", l.GetNamespace())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 4, len(row.Deltas))
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
	return resource.NewRoleBinding().NewInstance(k8sRB())
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
