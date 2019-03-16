package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
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
