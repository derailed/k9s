package model_test

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/watch"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

func TestTableRefresh(t *testing.T) {
	ta := model.NewTable(client.NewGVR("v1/pods"))
	ta.SetNamespace(client.NamespaceAll)

	l := tableListener{}
	ta.AddListener(&l)
	f := makeTableFactory()
	f.rows = []runtime.Object{mustLoad("p1")}
	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	ctx = context.WithValue(ctx, internal.KeyFields, "")
	ctx = context.WithValue(ctx, internal.KeyWithMetrics, false)
	ta.Refresh(ctx)
	data := ta.Peek()
	assert.Equal(t, 17, len(data.Header))
	assert.Equal(t, 1, len(data.RowEvents))
	assert.Equal(t, client.NamespaceAll, data.Namespace)
	assert.Equal(t, 1, l.count)
	assert.Equal(t, 0, l.errs)
}

func TestTableNS(t *testing.T) {
	ta := model.NewTable(client.NewGVR("v1/pods"))
	ta.SetNamespace("blee")

	assert.Equal(t, "blee", ta.GetNamespace())
	assert.False(t, ta.ClusterWide())
	assert.False(t, ta.InNamespace("zorg"))
}

func TestTableAddListener(t *testing.T) {
	ta := model.NewTable(client.NewGVR("v1/pods"))
	ta.SetNamespace("blee")

	assert.True(t, ta.Empty())
	l := tableListener{}
	ta.AddListener(&l)
}

func TestTableRmListener(t *testing.T) {
	ta := model.NewTable(client.NewGVR("v1/pods"))
	ta.SetNamespace("blee")

	l := tableListener{}
	ta.RemoveListener(&l)
}

// Helpers...

type tableListener struct {
	count, errs int
}

func (l *tableListener) TableDataChanged(render.TableData) {
	l.count++
}
func (l *tableListener) TableLoadFailed(error) {
	l.errs++
}

type tableFactory struct {
	rows []runtime.Object
}

var _ dao.Factory = tableFactory{}

func (f tableFactory) Client() client.Connection {
	return client.NewTestClient()
}
func (f tableFactory) Get(gvr, path string, wait bool, sel labels.Selector) (runtime.Object, error) {
	if len(f.rows) > 0 {
		return f.rows[0], nil
	}
	return nil, nil
}
func (f tableFactory) List(gvr, ns string, wait bool, sel labels.Selector) ([]runtime.Object, error) {
	if len(f.rows) > 0 {
		return f.rows, nil
	}
	return nil, nil
}
func (f tableFactory) ForResource(ns, gvr string) informers.GenericInformer {
	return nil
}
func (f tableFactory) CanForResource(ns, gvr string, verbs []string) (informers.GenericInformer, error) {
	return nil, nil
}
func (f tableFactory) WaitForCacheSync() {}
func (f tableFactory) Forwarders() watch.Forwarders {
	return nil
}
func (f tableFactory) DeleteForwarder(string) {}

func makeTableFactory() tableFactory {
	return tableFactory{}
}

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
