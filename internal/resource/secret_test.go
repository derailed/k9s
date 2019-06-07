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

func NewSecretListWithArgs(ns string, r *resource.Secret) resource.List {
	return resource.NewList(ns, "secret", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewSecretWithArgs(conn k8s.Connection, res resource.Cruder) *resource.Secret {
	r := &resource.Secret{Base: resource.NewBase(conn, res)}
	r.Factory = r
	return r
}

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
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sSecret(), nil)

	cm := NewSecretWithArgs(mc, mr)
	ma, err := cm.Get("blee/fred")

	assert.Nil(t, err)
	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Equal(t, cm.New(k8sSecret()), ma)
}

func TestSecretList(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List("blee")).ThenReturn(k8s.Collection{*k8sSecret()}, nil)

	cm := NewSecretWithArgs(mc, mr)
	ma, err := cm.List("blee")

	assert.Nil(t, err)
	mr.VerifyWasCalledOnce().List("blee")
	assert.Equal(t, resource.Columnars{cm.New(k8sSecret())}, ma)
}

func TestSecretDelete(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Delete("blee", "fred", true, true)).ThenReturn(nil)

	cm := NewSecretWithArgs(mc, mr)

	assert.Nil(t, cm.Delete("blee/fred", true, true))
	mr.VerifyWasCalledOnce().Delete("blee", "fred", true, true)
}

func TestSecretMarshal(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.Get("blee", "fred")).ThenReturn(k8sSecret(), nil)

	cm := NewSecretWithArgs(mc, mr)
	ma, err := cm.Marshal("blee/fred")

	mr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, secretYaml(), ma)
}

func TestSecretListHasName(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()

	l := NewSecretListWithArgs("blee", NewSecretWithArgs(mc, mr))
	assert.Equal(t, "secret", l.GetName())
}

func TestSecretListHasNamespace(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	l := NewSecretListWithArgs("blee", NewSecretWithArgs(mc, mr))
	assert.Equal(t, "blee", l.GetNamespace())
}

func TestSecretListHasResource(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	l := NewSecretListWithArgs("blee", NewSecretWithArgs(mc, mr))
	assert.NotNil(t, l.Resource())
}

func TestSecretListData(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockCruder()
	m.When(mr.List("blee")).ThenReturn(k8s.Collection{*k8sSecret()}, nil)

	l := NewSecretListWithArgs("blee", NewSecretWithArgs(mc, mr))
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
	assert.Equal(t, 4, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred", "Opaque", "2"}, row.Fields[:3])
}

// Helpers...

func newSecret() resource.Columnar {
	mc := NewMockConnection()
	return resource.NewSecret(mc).New(k8sSecret())
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
	return `apiVersion: v1
data:
  blee: YmxlZQ==
  duh: ZHVo
kind: Secret
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
type: Opaque
`
}
