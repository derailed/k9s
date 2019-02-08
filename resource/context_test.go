package resource_test

import (
	"testing"

	"github.com/derailed/k9s/resource"
	"github.com/derailed/k9s/resource/k8s"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/tools/clientcmd/api"
)

func TestCTXHeader(t *testing.T) {
	assert.Equal(t,
		resource.Row{"NAME", "CLUSTER", "AUTHINFO", "NAMESPACE"},
		newContext().Header(""),
	)
}

func TestCTXFieldsAllNS(t *testing.T) {
	r := newContext().Fields(resource.AllNamespaces)
	assert.Equal(t, "test", r[0])
	assert.Equal(t, "blee", r[1])
	assert.Equal(t, "secret", r[2])
}

func TestCTXSwitch(t *testing.T) {
	setup(t)

	ca := NewMockSwitchableRes()
	m.When(ca.Switch("fred")).ThenReturn(nil)

	ctx := resource.NewContextWithArgs(ca)
	err := ctx.Switch("fred")
	assert.Nil(t, err)
	ca.VerifyWasCalledOnce().Switch("fred")
}

func TestCTXList(t *testing.T) {
	setup(t)

	ca := NewMockSwitchableRes()
	m.When(ca.List("blee")).ThenReturn(k8s.Collection{*k8sNamedCTX()}, nil)

	ctx := resource.NewContextWithArgs(ca)
	cc, err := ctx.List("blee")
	assert.Nil(t, err)
	assert.Equal(t, resource.Columnars{ctx.NewInstance(k8sNamedCTX())}, cc)
	ca.VerifyWasCalledOnce().List("blee")
}
func TestCTXDelete(t *testing.T) {
	setup(t)

	ca := NewMockSwitchableRes()
	m.When(ca.Delete("", "fred")).ThenReturn(nil)

	cm := resource.NewContextWithArgs(ca)
	assert.Nil(t, cm.Delete("fred"))
	ca.VerifyWasCalledOnce().Delete("", "fred")
}

func TestCTXListSort(t *testing.T) {
	setup(t)

	ca := NewMockSwitchableRes()
	l := resource.NewContextListWithArgs("blee", resource.NewContextWithArgs(ca))
	kk := []string{"c", "b", "a"}
	l.SortFn()(kk)
	assert.Equal(t, []string{"a", "b", "c"}, kk)
}

func TestCTXListHasName(t *testing.T) {
	setup(t)

	ca := NewMockSwitchableRes()
	m.When(ca.List("blee")).ThenReturn(k8s.Collection{*k8sNamedCTX()}, nil)

	l := resource.NewContextListWithArgs("blee", resource.NewContextWithArgs(ca))
	assert.Equal(t, "ctx", l.GetName())
}

func TestCTXListHasNamespace(t *testing.T) {
	setup(t)

	ca := NewMockSwitchableRes()
	l := resource.NewContextListWithArgs("blee", resource.NewContextWithArgs(ca))
	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
}

func TestCTXListHasResource(t *testing.T) {
	setup(t)

	ca := NewMockSwitchableRes()
	l := resource.NewContextListWithArgs("blee", resource.NewContextWithArgs(ca))
	assert.NotNil(t, l.Resource())
}

func TestCTXListDescribe(t *testing.T) {
	setup(t)

	ca := NewMockSwitchableRes()
	m.When(ca.Get("blee", "fred")).ThenReturn(k8sNamedCTX(), nil)

	l := resource.NewContextListWithArgs("blee", resource.NewContextWithArgs(ca))
	props, err := l.Describe("blee/fred")
	assert.Nil(t, err)
	assert.Equal(t, 0, len(props))
	ca.VerifyWasCalledOnce().Get("blee", "fred")
}

func TestCTXListData(t *testing.T) {
	setup(t)

	ca := NewMockSwitchableRes()
	m.When(ca.List(resource.NotNamespaced)).ThenReturn(k8s.Collection{*k8sNamedCTX()}, nil)

	l := resource.NewContextListWithArgs("blee", resource.NewContextWithArgs(ca))
	// Make sure we can get deltas!
	for i := 0; i < 2; i++ {
		assert.Nil(t, l.Reconcile())
	}
	ca.VerifyWasCalled(m.Times(2)).List(resource.NotNamespaced)

	td := l.Data()
	assert.Equal(t, 1, len(td.Rows))
	assert.False(t, l.HasXRay())
	row := td.Rows["test"]
	assert.Equal(t, 4, len(row.Deltas))
	for _, d := range row.Deltas {
		assert.Equal(t, "", d)
	}
	assert.Equal(t, resource.Row{"test", "blee", "secret", ""}, row.Fields)
}

// Helpers...

func newContext() resource.Columnar {
	return resource.NewContext().NewInstance(k8sNamedCTX())
}

func k8sCTX() *api.Context {
	return &api.Context{
		LocationOfOrigin: "fred",
		Cluster:          "blee",
		AuthInfo:         "secret",
	}
}

func k8sNamedCTX() *k8s.NamedContext {
	ctx := k8s.NamedContext{
		Name: "test",
		Context: &api.Context{
			LocationOfOrigin: "fred",
			Cluster:          "blee",
			AuthInfo:         "secret",
		},
	}
	return &ctx
}
