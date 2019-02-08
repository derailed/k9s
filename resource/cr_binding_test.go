package resource_test

import (
	"testing"

	"github.com/derailed/k9s/resource"
	"github.com/derailed/k9s/resource/k8s"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCRBHeader(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "AGE"}, newCRB().Header(resource.DefaultNamespace))
}

func TestCRBHeaderAllNS(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "AGE"}, newCRB().Header(resource.AllNamespaces))
}

func TestCRBFields(t *testing.T) {
	r := newCRB().Fields(resource.AllNamespaces)
	assert.Equal(t, "fred", r[0])
}

func TestCRBMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sCRB(), nil)

	cm := resource.NewClusterRoleBindingWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, crbYaml(), ma)
}

func TestCRBListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sCRB()}, nil)

	l := resource.NewClusterRoleBindingListWithArgs("-", resource.NewClusterRoleBindingWithArgs(ca))
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

func newCRB() resource.Columnar {
	return resource.NewClusterRoleBinding().NewInstance(k8sCRB())
}

func crbYaml() string {
	return `typemeta:
  kind: ClusterRoleBinding
  apiversion: rbac.authorization.k8s.io/v1
objectmeta:
  name: fred
  generatename: ""
  namespace: blee
  selflink: ""
  uid: ""
  resourceversion: ""
  generation: 0
  creationtimestamp: "2018-12-14T10:36:43.326972-07:00"
  deletiontimestamp: null
  deletiongraceperiodseconds: null
  labels: {}
  annotations: {}
  ownerreferences: []
  initializers: null
  finalizers: []
  clustername: ""
  managedfields: []
subjects:
- kind: test
  apigroup: ""
  name: fred
  namespace: blee
roleref:
  apigroup: ""
  kind: ""
  name: ""
`
}
