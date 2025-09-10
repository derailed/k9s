// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"context"
	"testing"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestTableNew(t *testing.T) {
	v := ui.NewTable(client.NewGVR("fred"))
	v.Init(makeContext())

	assert.Equal(t, "fred", v.GVR().String())
}

func TestTableUpdate(t *testing.T) {
	v := ui.NewTable(client.NewGVR("fred"))
	v.Init(makeContext())

	data := makeTableData()
	cdata := v.Update(data, false)
	v.UpdateUI(cdata, data)

	assert.Equal(t, data.RowCount()+1, v.GetRowCount())
	assert.Equal(t, data.HeaderCount(), v.GetColumnCount())
}

func TestTableSelection(t *testing.T) {
	v := ui.NewTable(client.NewGVR("fred"))
	v.Init(makeContext())
	m := new(mockModel)
	v.SetModel(m)
	data := m.Peek()
	cdata := v.Update(data, false)
	v.UpdateUI(cdata, data)
	v.SelectRow(1, 0, true)

	r := v.GetSelectedRow("r1")
	if r != nil {
		assert.Equal(t, model1.Row{ID: "r1", Fields: model1.Fields{"blee", "duh", "fred"}}, *r)
	}
	assert.Equal(t, "r1", v.GetSelectedItem())
	assert.Equal(t, "blee", v.GetSelectedCell(0))
	assert.Equal(t, 1, v.GetSelectedRowIndex())
	assert.Equal(t, []string{"r1"}, v.GetSelectedItems())

	v.ClearSelection()
	v.SelectFirstRow()
	assert.Equal(t, 1, v.GetSelectedRowIndex())
}

// ----------------------------------------------------------------------------
// Helpers...

type mockModel struct{}

var _ ui.Tabular = &mockModel{}

func (*mockModel) SetViewSetting(context.Context, *config.ViewSetting) {}
func (*mockModel) SetInstance(string)                                  {}
func (*mockModel) SetLabelSelector(labels.Selector)                    {}
func (*mockModel) GetLabelSelector() labels.Selector                   { return nil }
func (*mockModel) Empty() bool                                         { return false }
func (*mockModel) RowCount() int                                       { return 1 }
func (*mockModel) HasMetrics() bool                                    { return true }
func (*mockModel) Peek() *model1.TableData                             { return makeTableData() }
func (*mockModel) Refresh(context.Context) error                       { return nil }
func (*mockModel) ClusterWide() bool                                   { return false }
func (*mockModel) GetNamespace() string                                { return "blee" }
func (*mockModel) SetNamespace(string)                                 {}
func (*mockModel) ToggleToast()                                        {}
func (*mockModel) AddListener(model.TableListener)                     {}
func (*mockModel) RemoveListener(model.TableListener)                  {}
func (*mockModel) Watch(context.Context) error                         { return nil }
func (*mockModel) Get(context.Context, string) (runtime.Object, error) { return nil, nil }
func (*mockModel) InNamespace(string) bool                             { return true }
func (*mockModel) SetRefreshRate(time.Duration)                        {}

func (*mockModel) Delete(context.Context, string, *metav1.DeletionPropagation, dao.Grace) error {
	return nil
}

func (*mockModel) Describe(context.Context, string) (string, error) {
	return "", nil
}

func (*mockModel) ToYAML(context.Context, string) (string, error) {
	return "", nil
}

func makeTableData() *model1.TableData {
	return model1.NewTableDataWithRows(
		client.NewGVR("test"),
		model1.Header{
			model1.HeaderColumn{Name: "A"},
			model1.HeaderColumn{Name: "B"},
			model1.HeaderColumn{Name: "C"},
		},
		model1.NewRowEventsWithEvts(
			model1.RowEvent{
				Row: model1.Row{
					ID:     "r1",
					Fields: model1.Fields{"blee", "duh", "fred"},
				},
			},
			model1.RowEvent{
				Row: model1.Row{
					ID:     "r2",
					Fields: model1.Fields{"blee", "duh", "zorg"},
				},
			},
		),
	)
}

func makeContext() context.Context {
	ctx := context.WithValue(context.Background(), internal.KeyStyles, config.NewStyles())
	ctx = context.WithValue(ctx, internal.KeyViewConfig, config.NewCustomView())

	return ctx
}
