package model_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/watch"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

func TestAliasList(t *testing.T) {
	a := model.Alias{}
	a.Init(render.ClusterScope, "aliases", makeFactory())

	ctx := context.WithValue(context.Background(), internal.KeyAliases, makeAliases())
	oo, err := a.List(ctx)

	assert.Nil(t, err)
	assert.Equal(t, 2, len(oo))
	assert.Equal(t, 2, len(oo[0].(render.AliasRes).Aliases))
}

func TestAliasHydrate(t *testing.T) {
	a := model.Alias{}
	a.Init(render.ClusterScope, "aliases", makeFactory())

	ctx := context.WithValue(context.Background(), internal.KeyAliases, makeAliases())
	oo, err := a.List(ctx)
	assert.Nil(t, err)

	rr := make(render.Rows, len(oo))
	assert.Nil(t, a.Hydrate(oo, rr, render.Alias{}))
	assert.Equal(t, 2, len(rr))
}

// ----------------------------------------------------------------------------
// Helpers...

func makeAliases() *dao.Alias {
	return &dao.Alias{
		Aliases: config.Aliases{
			Alias: config.Alias{
				"fred": "v1/fred",
				"f":    "v1/fred",
				"blee": "v1/blee",
				"b":    "v1/blee",
			},
		},
	}
}

type testFactory struct{}

var _ dao.Factory = testFactory{}

func (f testFactory) Client() client.Connection {
	return nil
}
func (f testFactory) Get(gvr, path string, wait bool, sel labels.Selector) (runtime.Object, error) {
	return nil, nil
}
func (f testFactory) List(gvr, ns string, wait bool, sel labels.Selector) ([]runtime.Object, error) {
	return nil, nil
}
func (f testFactory) ForResource(ns, gvr string) informers.GenericInformer {
	return nil
}
func (f testFactory) CanForResource(ns, gvr string, verbs []string) (informers.GenericInformer, error) {
	return nil, nil
}
func (f testFactory) WaitForCacheSync() {}
func (f testFactory) Forwarders() watch.Forwarders {
	return nil
}
func (f testFactory) DeleteForwarder(string) {}

func makeFactory() dao.Factory {
	return testFactory{}
}
