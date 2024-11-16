// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/watch"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

func TestTableReconcile(t *testing.T) {
	ta := NewTable(client.NewGVR("v1/pods"))
	ta.SetNamespace(client.NamespaceAll)

	f := makeFactory()
	f.rows = []runtime.Object{load(t, "p1")}
	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	ctx = context.WithValue(ctx, internal.KeyFields, "")
	ctx = context.WithValue(ctx, internal.KeyWithMetrics, false)
	err := ta.reconcile(ctx)
	assert.Nil(t, err)
	data := ta.Peek()
	assert.Equal(t, 25, data.HeaderCount())
	assert.Equal(t, 1, data.RowCount())
	assert.Equal(t, client.NamespaceAll, data.GetNamespace())
}

func TestTableList(t *testing.T) {
	ta := NewTable(client.NewGVR("v1/pods"))
	ta.SetNamespace("blee")

	acc := accessor{}
	ctx := context.WithValue(context.Background(), internal.KeyFactory, makeFactory())
	rows, err := ta.list(ctx, &acc)
	assert.Nil(t, err)
	assert.Equal(t, 1, len(rows))
}

func TestTableGet(t *testing.T) {
	ta := NewTable(client.NewGVR("v1/pods"))
	ta.SetNamespace("blee")

	f := makeFactory()
	f.rows = []runtime.Object{load(t, "p1")}
	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	ctx = context.WithValue(ctx, internal.KeyWithMetrics, false)
	row, err := ta.Get(ctx, "fred")
	assert.Nil(t, err)
	assert.NotNil(t, row)
	assert.Equal(t, 5, len(row.(*render.PodWithMetrics).Raw.Object))
}

func TestTableMeta(t *testing.T) {
	uu := map[string]struct {
		gvr      string
		accessor dao.Accessor
		renderer model1.Renderer
	}{
		"generic": {
			gvr:      "containers",
			accessor: &dao.Container{},
			renderer: &render.Container{},
		},
		"node": {
			gvr:      "v1/nodes",
			accessor: &dao.Node{},
			renderer: &render.Node{},
		},
		"table": {
			gvr:      "v1/events",
			accessor: &dao.Table{},
			renderer: &render.Event{},
		},
	}

	for k := range uu {
		u := uu[k]
		ta := NewTable(client.NewGVR(u.gvr))
		m := resourceMeta(ta.gvr)

		assert.Equal(t, u.accessor, m.DAO)
		assert.Equal(t, u.renderer, m.Renderer)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func mustLoad(n string) *unstructured.Unstructured {
	raw, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", n))
	if err != nil {
		panic(err)
	}
	var o unstructured.Unstructured
	if err = json.Unmarshal(raw, &o); err != nil {
		panic(err)
	}
	return &o
}

func load(t *testing.T, n string) *unstructured.Unstructured {
	raw, err := os.ReadFile(fmt.Sprintf("testdata/%s.json", n))
	assert.Nil(t, err)
	var o unstructured.Unstructured
	err = json.Unmarshal(raw, &o)
	assert.Nil(t, err)
	return &o
}

// ----------------------------------------------------------------------------

func makeFactory() testFactory {
	return testFactory{}
}

type testFactory struct {
	rows []runtime.Object
}

var _ dao.Factory = testFactory{}

func (f testFactory) Client() client.Connection {
	return client.NewTestAPIClient()
}

func (f testFactory) Get(gvr, path string, wait bool, sel labels.Selector) (runtime.Object, error) {
	if len(f.rows) > 0 {
		return f.rows[0], nil
	}
	return nil, nil
}

func (f testFactory) List(gvr, ns string, wait bool, sel labels.Selector) ([]runtime.Object, error) {
	if len(f.rows) > 0 {
		return f.rows, nil
	}
	return nil, nil
}

func (f testFactory) ForResource(ns, gvr string) (informers.GenericInformer, error) {
	return nil, nil
}

func (f testFactory) CanForResource(ns, gvr string, verbs []string) (informers.GenericInformer, error) {
	return nil, nil
}
func (f testFactory) WaitForCacheSync() {}
func (f testFactory) Forwarders() watch.Forwarders {
	return nil
}
func (f testFactory) DeleteForwarder(string) {}

// ----------------------------------------------------------------------------

type accessor struct {
	gvr client.GVR
}

var _ dao.Accessor = (*accessor)(nil)

func (a *accessor) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	return []runtime.Object{&render.PodWithMetrics{Raw: mustLoad("p1")}}, nil
}

func (a *accessor) Get(ctx context.Context, path string) (runtime.Object, error) {
	return &render.PodWithMetrics{Raw: mustLoad("p1")}, nil
}

func (a *accessor) Init(_ dao.Factory, gvr client.GVR) {
	a.gvr = gvr
}

func (a *accessor) GVR() string {
	return a.gvr.String()
}
