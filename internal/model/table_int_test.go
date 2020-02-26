package model

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/derailed/k9s/internal"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/watch"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
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
	assert.Equal(t, 17, len(data.Header))
	assert.Equal(t, 1, len(data.RowEvents))
	assert.Equal(t, client.NamespaceAll, data.Namespace)
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
	pd := dao.Pod{}
	pd.Init(makeFactory(), client.NewGVR("v1/pods"))
	uu := map[string]struct {
		gvr      string
		accessor dao.Accessor
		renderer Renderer
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
			gvr:      "v1/configmaps",
			accessor: &dao.Table{},
			renderer: &render.Generic{},
		},
	}

	for k := range uu {
		u := uu[k]
		ta := NewTable(client.NewGVR(u.gvr))
		m := ta.resourceMeta()

		assert.Equal(t, u.accessor, m.DAO)
		assert.Equal(t, u.renderer, m.Renderer)
	}
}

func TestTableHydrate(t *testing.T) {
	oo := []runtime.Object{
		&render.PodWithMetrics{Raw: load(t, "p1")},
	}
	rr := make([]render.Row, 1)

	assert.Nil(t, hydrate("blee", oo, rr, render.Pod{}))
	assert.Equal(t, 1, len(rr))
	assert.Equal(t, 17, len(rr[0].Fields))
}

func TestTableGenericHydrate(t *testing.T) {
	raw := raw(t, "p1")
	tt := metav1beta1.Table{
		ColumnDefinitions: []metav1beta1.TableColumnDefinition{
			{Name: "c1"},
			{Name: "c2"},
		},
		Rows: []metav1beta1.TableRow{
			{
				Cells:  []interface{}{"fred", 10},
				Object: runtime.RawExtension{Raw: raw},
			},
			{
				Cells:  []interface{}{"blee", 20},
				Object: runtime.RawExtension{Raw: raw},
			},
		},
	}
	rr := make([]render.Row, 2)
	re := render.Generic{}
	re.SetTable(&tt)

	assert.Nil(t, genericHydrate("blee", &tt, rr, &re))
	assert.Equal(t, 2, len(rr))
	assert.Equal(t, 3, len(rr[0].Fields))
}

// ----------------------------------------------------------------------------
// Helpers...

func mustLoad(n string) *unstructured.Unstructured {
	raw, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s.json", n))
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
	raw, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s.json", n))
	assert.Nil(t, err)
	var o unstructured.Unstructured
	err = json.Unmarshal(raw, &o)
	assert.Nil(t, err)
	return &o
}

func raw(t *testing.T, n string) []byte {
	raw, err := ioutil.ReadFile(fmt.Sprintf("testdata/%s.json", n))
	assert.Nil(t, err)
	return raw
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
	return client.NewTestClient()
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
