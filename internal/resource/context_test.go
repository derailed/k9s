package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	api "k8s.io/client-go/tools/clientcmd/api"
)

func NewContextListWithArgs(ns string, ctx *resource.Context) resource.List {
	return resource.NewList(resource.NotNamespaced, "ctx", ctx, resource.SwitchAccess)
}

func NewContextWithArgs(c k8s.Connection, s resource.SwitchableResource) *resource.Context {
	ctx := &resource.Context{Base: resource.NewBase(c, s)}
	ctx.Factory = ctx
	return ctx
}

func TestCTXSwitch(t *testing.T) {
	conn := NewMockConnection()
	ca := NewMockSwitchableResource()
	m.When(ca.Switch("fred")).ThenReturn(nil)

	ctx := NewContextWithArgs(conn, ca)
	err := ctx.Switch("fred")

	assert.Nil(t, err)
	ca.VerifyWasCalledOnce().Switch("fred")
}

func TestCTXList(t *testing.T) {
	conn := NewMockConnection()
	ca := NewMockSwitchableResource()
	m.When(ca.List("blee")).ThenReturn(k8s.Collection{*k8sNamedCTX()}, nil)

	ctx := NewContextWithArgs(conn, ca)
	cc, err := ctx.List("blee")

	assert.Nil(t, err)
	assert.Equal(t, resource.Columnars{ctx.New(k8sNamedCTX())}, cc)
	ca.VerifyWasCalledOnce().List("blee")
}

func TestCTXDelete(t *testing.T) {
	conn := NewMockConnection()
	ca := NewMockSwitchableResource()
	m.When(ca.Delete("", "fred")).ThenReturn(nil)

	ctx := NewContextWithArgs(conn, ca)

	assert.Nil(t, ctx.Delete("fred"))
	ca.VerifyWasCalledOnce().Delete("", "fred")
}

func TestCTXListHasName(t *testing.T) {
	conn := NewMockConnection()
	ca := NewMockSwitchableResource()

	ctx := NewContextWithArgs(conn, ca)
	l := NewContextListWithArgs("blee", ctx)

	assert.Equal(t, "ctx", l.GetName())
}

func TestCTXListHasNamespace(t *testing.T) {
	conn := NewMockConnection()
	ca := NewMockSwitchableResource()

	ctx := NewContextWithArgs(conn, ca)
	l := NewContextListWithArgs("blee", ctx)

	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
}

func TestCTXListHasResource(t *testing.T) {
	conn := NewMockConnection()
	ca := NewMockSwitchableResource()

	ctx := NewContextWithArgs(conn, ca)
	l := NewContextListWithArgs("blee", ctx)

	assert.NotNil(t, l.Resource())
}

// Helpers...

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
