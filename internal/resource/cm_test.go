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

func NewConfigMapListWithArgs(ns string, r *resource.ConfigMap) resource.List {
	return resource.NewList(ns, "cm", r, resource.AllVerbsAccess|resource.DescribeAccess)
}

func NewConfigMapWithArgs(rc k8s.Connection, res resource.Cruder) *resource.ConfigMap {
	r := &resource.ConfigMap{Base: resource.NewBase(rc, res)}
	r.Factory = r
	return r
}

func TestCMFieldsAllNS(t *testing.T) {
	r := newConfigMap().Fields(resource.AllNamespaces)
	assert.Equal(t, "blee", r[0])
	assert.Equal(t, "fred", r[1])
	assert.Equal(t, "2", r[2])
}

func TestCMFields(t *testing.T) {
	r := newConfigMap().Fields("blee")
	assert.Equal(t, "fred", r[0])
	assert.Equal(t, "2", r[1])
}

func TestCMGet(t *testing.T) {
	rc := NewMockConnection()
	cr := NewMockCruder()
	m.When(cr.Get("blee", "fred")).ThenReturn(k8sCM(), nil)

	cm := NewConfigMapWithArgs(rc, cr)
	ma, err := cm.Get("blee/fred")

	assert.Nil(t, err)
	cr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Equal(t, cm.New(k8sCM()), ma)
}

func TestCMList(t *testing.T) {
	rc := NewMockConnection()
	cr := NewMockCruder()
	m.When(cr.List("blee")).ThenReturn(k8s.Collection{*k8sCM()}, nil)

	cm := NewConfigMapWithArgs(rc, cr)
	ma, err := cm.List("blee")

	assert.Nil(t, err)
	cr.VerifyWasCalledOnce().List("blee")
	assert.Equal(t, resource.Columnars{cm.New(k8sCM())}, ma)
}

func TestCMDelete(t *testing.T) {
	rc := NewMockConnection()
	cr := NewMockCruder()
	m.When(cr.Delete("blee", "fred")).ThenReturn(nil)

	cm := NewConfigMapWithArgs(rc, cr)

	assert.Nil(t, cm.Delete("blee/fred"))
	cr.VerifyWasCalledOnce().Delete("blee", "fred")
}

func TestCMMarshal(t *testing.T) {
	rc := NewMockConnection()
	cr := NewMockCruder()
	m.When(cr.Get("blee", "fred")).ThenReturn(k8sCM(), nil)

	cm := NewConfigMapWithArgs(rc, cr)
	ma, err := cm.Marshal("blee/fred")

	cr.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, cmYaml(), ma)
}

func TestCMListHasName(t *testing.T) {
	rc := NewMockConnection()
	cr := NewMockCruder()

	cm := NewConfigMapWithArgs(rc, cr)
	l := NewConfigMapListWithArgs("blee", cm)

	assert.Equal(t, "cm", l.GetName())
}

func TestCMListHasNamespace(t *testing.T) {
	rc := NewMockConnection()
	cr := NewMockCruder()

	cm := NewConfigMapWithArgs(rc, cr)
	l := NewConfigMapListWithArgs("blee", cm)

	assert.Equal(t, "blee", l.GetNamespace())
}

func TestCMListHasResource(t *testing.T) {
	rc := NewMockConnection()
	cr := NewMockCruder()

	cm := NewConfigMapWithArgs(rc, cr)
	l := NewConfigMapListWithArgs("blee", cm)

	assert.NotNil(t, l.Resource())
}

func TestCMListData(t *testing.T) {
	rc := NewMockConnection()
	cr := NewMockCruder()
	m.When(cr.List("blee")).ThenReturn(k8s.Collection{*k8sCM()}, nil)

	cm := NewConfigMapWithArgs(rc, cr)
	l := NewConfigMapListWithArgs("blee", cm)

	// Make sure we crn get deltas!
	for i := 0; i < 2; i++ {
		err := l.Reconcile(nil, nil)
		assert.Nil(t, err)
	}

	cr.VerifyWasCalled(m.Times(2)).List("blee")

	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))

	assert.Equal(t, "blee", l.GetNamespace())
	row := td.Rows["blee/fred"]
	assert.Equal(t, 3, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred", "2"}, row.Fields[:2])
}

// Helpers...

func newConfigMap() resource.Columnar {
	rc := NewMockConnection()
	return resource.NewConfigMap(rc).New(k8sCM())
}

func k8sCM() *v1.ConfigMap {
	return &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:              "fred",
			Namespace:         "blee",
			CreationTimestamp: metav1.Time{testTime()},
		},
		Data: map[string]string{"blee": "blee", "duh": "duh"},
	}
}

func cmYaml() string {
	return `apiVersion: v1
data:
  blee: blee
  duh: duh
kind: ConfigMap
metadata:
  creationTimestamp: "2018-12-14T17:36:43Z"
  name: fred
  namespace: blee
`
}
