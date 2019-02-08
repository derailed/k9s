package resource_test

import (
	"strings"
	"testing"
	"time"

	"github.com/derailed/k9s/resource"
	"github.com/derailed/k9s/resource/k8s"
	m "github.com/petergtz/pegomock"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCRListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewClusterRoleList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	assert.Equal(t, "clusterrole", l.GetName())
	for _, a := range []int{resource.GetAccess, resource.ListAccess, resource.DeleteAccess, resource.ViewAccess, resource.EditAccess} {
		assert.True(t, l.Access(a))
	}
}

func TestCRHeader(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "AGE"}, newClusterRole().Header(resource.DefaultNamespace))
}

func TestCRHeaderAllNS(t *testing.T) {
	assert.Equal(t, resource.Row{"NAME", "AGE"}, newClusterRole().Header(resource.AllNamespaces))
}

func TestCRFields(t *testing.T) {
	r := newClusterRole().Fields("blee")
	assert.Equal(t, "fred"+strings.Repeat(" ", 66), r[0])
}

func TestCRFieldsAllNS(t *testing.T) {
	r := newClusterRole().Fields(resource.AllNamespaces)
	assert.Equal(t, "fred"+strings.Repeat(" ", 66), r[0])
}

func TestCRMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sCR(), nil)

	cm := resource.NewClusterRoleWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, crYaml(), ma)
}

func TestCRListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sCR()}, nil)

	l := resource.NewClusterRoleListWithArgs("-", resource.NewClusterRoleWithArgs(ca))
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
	assert.Equal(t, resource.Row{"fred" + strings.Repeat(" ", 66)}, row.Fields[:1])
}

func TestCRListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sCR(), nil)
	l := resource.NewClusterRoleListWithArgs("blee", resource.NewClusterRoleWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
}

// Helpers...

func k8sCR() *rbacv1.ClusterRole {
	return &rbacv1.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "fred",
			Namespace:         "blee",
			CreationTimestamp: metav1.Time{testTime()},
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
	return resource.NewClusterRole().NewInstance(k8sCR())
}

func testTime() time.Time {
	t, err := time.Parse(time.RFC3339, "2018-12-14T10:36:43.326972-07:00")
	if err != nil {
		log.Println("TestTime Failed", err)
	}
	return t
}

func crYaml() string {
	return `typemeta:
  kind: ClusterRole
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
rules:
- verbs:
  - get
  - list
  apigroups:
  - ""
  resources: []
  resourcenames:
  - pod
  nonresourceurls: []
aggregationrule: null
`
}
