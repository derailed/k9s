// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/watch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
)

func TestTableRefresh(t *testing.T) {
	ta := model.NewTable(client.PodGVR)
	ta.SetNamespace(client.NamespaceAll)

	l := tableListener{}
	ta.AddListener(&l)
	f := makeTableFactory()
	f.rows = []runtime.Object{mustLoad("p1")}
	ctx := context.WithValue(context.Background(), internal.KeyFactory, f)
	ctx = context.WithValue(ctx, internal.KeyFields, "")
	ctx = context.WithValue(ctx, internal.KeyWithMetrics, false)
	require.NoError(t, ta.Refresh(ctx))
	data := ta.Peek()
	assert.Equal(t, 26, data.HeaderCount())
	assert.Equal(t, 1, data.RowCount())
	assert.Equal(t, client.NamespaceAll, data.GetNamespace())
	assert.Equal(t, 1, l.count)
	assert.Equal(t, 0, l.errs)
}

func TestTableNS(t *testing.T) {
	ta := model.NewTable(client.PodGVR)
	ta.SetNamespace("blee")

	assert.Equal(t, "blee", ta.GetNamespace())
	assert.False(t, ta.ClusterWide())
	assert.False(t, ta.InNamespace("zorg"))
}

func TestTableAddListener(t *testing.T) {
	ta := model.NewTable(client.PodGVR)
	ta.SetNamespace("blee")

	assert.True(t, ta.Empty())
	l := tableListener{}
	ta.AddListener(&l)
}

func TestTableRmListener(*testing.T) {
	ta := model.NewTable(client.PodGVR)
	ta.SetNamespace("blee")

	l := tableListener{}
	ta.RemoveListener(&l)
}

// Helpers...

type tableListener struct {
	count, errs int
}

func (*tableListener) TableNoData(*model1.TableData) {}

func (l *tableListener) TableDataChanged(*model1.TableData) {
	l.count++
}

func (l *tableListener) TableLoadFailed(error) {
	l.errs++
}

type tableFactory struct {
	rows []runtime.Object
}

var _ dao.Factory = tableFactory{}

func (tableFactory) Client() client.Connection {
	return client.NewTestAPIClient()
}

func (f tableFactory) Get(*client.GVR, string, bool, labels.Selector) (runtime.Object, error) {
	if len(f.rows) > 0 {
		return f.rows[0], nil
	}
	return nil, nil
}

func (f tableFactory) List(*client.GVR, string, bool, labels.Selector) ([]runtime.Object, error) {
	if len(f.rows) > 0 {
		return f.rows, nil
	}
	return nil, nil
}

func (tableFactory) ForResource(string, *client.GVR) (informers.GenericInformer, error) {
	return nil, nil
}

func (tableFactory) CanForResource(string, *client.GVR, []string) (informers.GenericInformer, error) {
	return nil, nil
}
func (tableFactory) WaitForCacheSync() {}
func (tableFactory) Forwarders() watch.Forwarders {
	return nil
}
func (tableFactory) DeleteForwarder(string) {}

func makeTableFactory() tableFactory {
	return tableFactory{}
}

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
