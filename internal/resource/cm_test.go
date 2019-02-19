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

func TestCMHeader(t *testing.T) {
	assert.Equal(t,
		resource.Row{"NAME", "DATA", "AGE"},
		newConfigMap().Header(resource.DefaultNamespace),
	)
}

func TestCMHeaderAllNS(t *testing.T) {
	assert.Equal(t,
		resource.Row{"NAMESPACE", "NAME", "DATA", "AGE"},
		newConfigMap().Header(resource.AllNamespaces),
	)
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
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sCM(), nil)

	cm := resource.NewConfigMapWithArgs(ca)
	ma, err := cm.Get("blee/fred")
	assert.Nil(t, err)
	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Equal(t, cm.NewInstance(k8sCM()), ma)
}

func TestCMList(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List("blee")).ThenReturn(k8s.Collection{*k8sCM()}, nil)

	cm := resource.NewConfigMapWithArgs(ca)
	ma, err := cm.List("blee")
	assert.Nil(t, err)
	ca.VerifyWasCalledOnce().List("blee")
	assert.Equal(t, resource.Columnars{cm.NewInstance(k8sCM())}, ma)
}

func TestCMDelete(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Delete("blee", "fred")).ThenReturn(nil)

	cm := resource.NewConfigMapWithArgs(ca)
	assert.Nil(t, cm.Delete("blee/fred"))
	ca.VerifyWasCalledOnce().Delete("blee", "fred")
}

func TestCMMarshal(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sCM(), nil)

	cm := resource.NewConfigMapWithArgs(ca)
	ma, err := cm.Marshal("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, cmYaml(), ma)
}

func TestCMListSort(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	l := resource.NewConfigMapListWithArgs("blee", resource.NewConfigMapWithArgs(ca))
	kk := []string{"c", "b", "a"}
	l.SortFn()(kk)
	assert.Equal(t, []string{"a", "b", "c"}, kk)
}

func TestCMListHasName(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	l := resource.NewConfigMapListWithArgs("blee", resource.NewConfigMapWithArgs(ca))
	assert.Equal(t, "cm", l.GetName())
}

func TestCMListHasNamespace(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	l := resource.NewConfigMapListWithArgs("blee", resource.NewConfigMapWithArgs(ca))
	assert.Equal(t, "blee", l.GetNamespace())
}

func TestCMListHasResource(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	l := resource.NewConfigMapListWithArgs("blee", resource.NewConfigMapWithArgs(ca))
	assert.NotNil(t, l.Resource())
}

func TestCMListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sCM(), nil)

	l := resource.NewConfigMapListWithArgs("blee", resource.NewConfigMapWithArgs(ca))
	props, err := l.Describe("blee/fred")

	ca.VerifyWasCalledOnce().Get("blee", "fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
}

func TestCMListData(t *testing.T) {
	setup(t)

	ca := NewMockCaller()
	m.When(ca.List("blee")).ThenReturn(k8s.Collection{*k8sCM()}, nil)

	l := resource.NewConfigMapListWithArgs("blee", resource.NewConfigMapWithArgs(ca))
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
	assert.Equal(t, 3, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"fred", "2"}, row.Fields[:2])
}

// Helpers...

func newConfigMap() resource.Columnar {
	return resource.NewConfigMap().NewInstance(k8sCM())
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
