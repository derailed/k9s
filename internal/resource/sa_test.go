package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/k8s"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestSaListAccess(t *testing.T) {
	ns := "blee"
	l := resource.NewServiceAccountList(resource.AllNamespaces)
	l.SetNamespace(ns)

	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
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
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sSA(), nil)

	cm := resource.NewServiceAccountWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, saYaml(), ma)
}

func TestSAListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List("-")).ThenReturn(k8s.Collection{*k8sSA()}, nil)

	l := resource.NewServiceAccountListWithArgs("-", resource.NewServiceAccountWithArgs(ca))
	// Make sure we can get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile()
		assert.Nil(t, err)
	}

	ca.VerifyWasCalled(m.Times(2)).List("-")
	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
	assert.False(t, l.HasXRay())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 3, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred"}, row.Fields[:1])
}

func TestSAListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sSA(), nil)
	l := resource.NewServiceAccountListWithArgs("blee", resource.NewServiceAccountWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
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
	return resource.NewServiceAccount().NewInstance(k8sSA())
}

func saHeader() resource.Row {
	return resource.Row{"NAME", "SECRET", "AGE"}
}

func saYaml() string {
	return `typemeta:
  kind: ServiceAccount
  apiversion: v1
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
secrets:
- kind: ""
  namespace: ""
  name: blee
  uid: ""
  apiversion: ""
  resourceversion: ""
  fieldpath: ""
imagepullsecrets: []
automountserviceaccounttoken: null
`
}
