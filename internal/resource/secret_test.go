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

func TestSecretHeader(t *testing.T) {
	assert.Equal(t,
		resource.Row{"NAME", "TYPE", "DATA", "AGE"},
		newSecret().Header(resource.DefaultNamespace),
	)
}

func TestSecretHeaderAllNS(t *testing.T) {
	assert.Equal(t,
		resource.Row{"NAMESPACE", "NAME", "TYPE", "DATA", "AGE"},
		newSecret().Header(resource.AllNamespaces),
	)
}

func TestSecretFieldsAllNS(t *testing.T) {
	r := newSecret().Fields(resource.AllNamespaces)
	assert.Equal(t, "blee", r[0])
	assert.Equal(t, "fred", r[1])
	assert.Equal(t, "Opaque", r[2])
	assert.Equal(t, "2", r[3])
}

func TestSecretFields(t *testing.T) {
	r := newSecret().Fields("blee")
	assert.Equal(t, "fred", r[0])
	assert.Equal(t, "Opaque", r[1])
	assert.Equal(t, "2", r[2])
}

func TestSecretGet(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sSecret(), nil)

	cm := resource.NewSecretWithArgs(ca)
	ma, err := cm.Get("blee/fred")
	assert.Nil(t, err)
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Equal(t, cm.NewInstance(k8sSecret()), ma)
}

func TestSecretList(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List("blee")).ThenReturn(k8s.Collection{*k8sSecret()}, nil)

	cm := resource.NewSecretWithArgs(ca)
	ma, err := cm.List("blee")
	assert.Nil(t, err)
	ca.VerifyWasCalledOnce().List("blee")
	assert.Equal(t, resource.Columnars{cm.NewInstance(k8sSecret())}, ma)
}

func TestSecretDelete(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Delete("blee", "fred")).ThenReturn(nil)

	cm := resource.NewSecretWithArgs(ca)
	assert.Nil(t, cm.Delete("blee/fred"))
	ca.VerifyWasCalledOnce().Delete("blee", "fred")
}

func TestSecretMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sSecret(), nil)

	cm := resource.NewSecretWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, secretYaml(), ma)
}

func TestSecretListSort(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	l := resource.NewSecretListWithArgs("blee", resource.NewSecretWithArgs(ca))
	kk := []string{"c", "b", "a"}
	l.SortFn()(kk)
	assert.Equal(t, []string{"a", "b", "c"}, kk)
}

func TestSecretListHasName(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	l := resource.NewSecretListWithArgs("blee", resource.NewSecretWithArgs(ca))
	assert.Equal(t, "secret", l.GetName())
}

func TestSecretListHasNamespace(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	l := resource.NewSecretListWithArgs("blee", resource.NewSecretWithArgs(ca))
	assert.Equal(t, "blee", l.GetNamespace())
}

func TestSecretListHasResource(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	l := resource.NewSecretListWithArgs("blee", resource.NewSecretWithArgs(ca))
	assert.NotNil(t, l.Resource())
}

func TestSecretListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sSecret(), nil)

	l := resource.NewSecretListWithArgs("blee", resource.NewSecretWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
}

func TestSecretListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List("blee")).ThenReturn(k8s.Collection{*k8sSecret()}, nil)

	l := resource.NewSecretListWithArgs("blee", resource.NewSecretWithArgs(ca))
	// Make sure we can get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile()
		assert.Nil(t, err)
	}

	ca.VerifyWasCalled(m.Times(2)).List("blee")

	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))

	assert.Equal(t, "blee", l.GetNamespace())
	assert.False(t, l.HasXRay())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 4, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred", "Opaque", "2"}, row.Fields[:3])
}

// Helpers...

func newSecret() resource.Columnar {
	return resource.NewSecret().NewInstance(k8sSecret())
}

func k8sSecret() *v1.Secret {
	secrets := map[string]string{"blee": "blee", "duh": "duh"}
	data := map[string][]byte{}
	for k, v := range secrets {
		data[k] = []byte(v)
	}
	return &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "fred",
			Namespace:         "blee",
			CreationTimestamp: metav1.Time{testTime()},
		},
		Type: v1.SecretTypeOpaque,
		Data: data,
	}
}

func secretYaml() string {
	return `typemeta:
  kind: Secret
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
data:
  blee:
  - 98
  - 108
  - 101
  - 101
  duh:
  - 100
  - 117
  - 104
stringdata: {}
type: Opaque
`
}
