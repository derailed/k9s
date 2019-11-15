package resource_test

import (
	"testing"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	api "k8s.io/client-go/tools/clientcmd/api"
)

func NewContextListWithArgs(ns string, ctx *resource.Context) resource.List {
	return resource.NewList(resource.NotNamespaced, "ctx", ctx, resource.SwitchAccess)
}

func NewContextWithArgs(c k8s.Connection, s resource.SwitchableCruder) *resource.Context {
	ctx := &resource.Context{Base: resource.NewBase(c, s)}
	ctx.Factory = ctx
	return ctx
}

func TestCTXSwitch(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockSwitchableCruder()
	m.When(mr.Switch("fred")).ThenReturn(nil)

	ctx := NewContextWithArgs(mc, mr)
	err := ctx.Switch("fred")

	assert.Nil(t, err)
	mr.VerifyWasCalledOnce().Switch("fred")
}

func TestCTXList(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockSwitchableCruder()
	m.When(mr.List("blee", metav1.ListOptions{})).ThenReturn(k8s.Collection{*k8sNamedCTX()}, nil)

	ctx := NewContextWithArgs(mc, mr)
	cc, err := ctx.List("blee", metav1.ListOptions{})

	assert.Nil(t, err)
	assert.Equal(t, resource.Columnars{ctx.New(k8sNamedCTX())}, cc)
	mr.VerifyWasCalledOnce().List("blee", metav1.ListOptions{})
}

func TestCTXDelete(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockSwitchableCruder()
	m.When(mr.Delete("", "fred", true, true)).ThenReturn(nil)

	ctx := NewContextWithArgs(mc, mr)

	assert.Nil(t, ctx.Delete("fred", true, true))
	mr.VerifyWasCalledOnce().Delete("", "fred", true, true)
}

func TestCTXListHasName(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockSwitchableCruder()

	ctx := NewContextWithArgs(mc, mr)
	l := NewContextListWithArgs("blee", ctx)

	assert.Equal(t, "ctx", l.GetName())
}

func TestCTXListHasNamespace(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockSwitchableCruder()

	ctx := NewContextWithArgs(mc, mr)
	l := NewContextListWithArgs("blee", ctx)

	assert.Equal(t, resource.NotNamespaced, l.GetNamespace())
}

func TestCTXListHasResource(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockSwitchableCruder()

	ctx := NewContextWithArgs(mc, mr)
	l := NewContextListWithArgs("blee", ctx)

	assert.NotNil(t, l.Resource())
}

func TestCTXHeader(t *testing.T) {
	mc := NewMockConnection()
	mr := NewMockSwitchableCruder()

	ctx := NewContextWithArgs(mc, mr)

	assert.Equal(t, 4, len(ctx.Header("")))
}

func TestCTXFields(t *testing.T) {
	mc := NewMockConnection()
	m.When(mc.Config()).ThenReturn(k8sConfig())
	mr := NewMockSwitchableCruder()
	m.When(mr.MustCurrentContextName()).ThenReturn("test")

	ctx := NewContextWithArgs(mc, mr)
	c := ctx.New(k8sNamedCTX())

	assert.Equal(t, 4, len(c.Fields("")))
	assert.Equal(t, "test*", c.Fields("")[0])
}

// Helpers...

func k8sConfig() *k8s.Config {
	ctx := "test"
	f := genericclioptions.ConfigFlags{
		Context: &ctx,
	}
	return k8s.NewConfig(&f)
}

func k8sNamedCTX() *k8s.NamedContext {
	return k8s.NewNamedContext(
		k8sConfig(),
		"test",
		&api.Context{
			LocationOfOrigin: "fred",
			Cluster:          "blee",
			AuthInfo:         "secret",
		},
	)
}
