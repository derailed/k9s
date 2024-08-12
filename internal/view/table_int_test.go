// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"testing"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestTableSave(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	assert.NoError(t, v.Init(makeContext()))
	v.SetTitle("k9s-test")

	assert.NoError(t, ensureDumpDir("/tmp/test-dumps"))
	dir := v.app.Config.K9s.ContextScreenDumpDir()
	c1, _ := os.ReadDir(dir)
	v.saveCmd(nil)

	c2, _ := os.ReadDir(dir)
	assert.Equal(t, len(c2), len(c1)+1)
}

func TestTableNew(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	assert.NoError(t, v.Init(makeContext()))

	data := model1.NewTableDataWithRows(
		client.NewGVR("test"),
		model1.Header{
			model1.HeaderColumn{Name: "NAMESPACE"},
			model1.HeaderColumn{Name: "NAME", Align: tview.AlignRight},
			model1.HeaderColumn{Name: "FRED"},
			model1.HeaderColumn{Name: "AGE", Time: true, Decorator: render.AgeDecorator},
		},
		model1.NewRowEventsWithEvts(
			model1.RowEvent{
				Row: model1.Row{
					Fields: model1.Fields{"ns1", "a", "10", "3m"},
				},
			},
			model1.RowEvent{
				Row: model1.Row{
					Fields: model1.Fields{"ns1", "b", "15", "1m"},
				},
			},
		),
	)
	cdata := v.Update(data, false)
	v.UpdateUI(cdata, data)

	assert.Equal(t, 3, v.GetRowCount())
}

func TestTableViewFilter(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	assert.NoError(t, v.Init(makeContext()))
	v.SetModel(&mockTableModel{})
	v.Refresh()

	v.CmdBuff().SetActive(true)
	v.CmdBuff().SetText("blee", "")

	assert.Equal(t, 5, v.GetRowCount())
}

func TestTableViewSort(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	assert.NoError(t, v.Init(makeContext()))
	v.SetModel(new(mockTableModel))

	uu := map[string]struct {
		sortCol  string
		sorted   []string
		reversed []string
	}{
		"by_name": {
			sortCol:  "NAME",
			sorted:   []string{"r0", "r1", "r2", "r3"},
			reversed: []string{"r3", "r2", "r1", "r0"},
		},
		"by_age": {
			sortCol:  "AGE",
			sorted:   []string{"r0", "r1", "r2", "r3"},
			reversed: []string{"r3", "r2", "r1", "r0"},
		},
		"by_fred": {
			sortCol:  "FRED",
			sorted:   []string{"r3", "r2", "r0", "r1"},
			reversed: []string{"r1", "r0", "r2", "r3"},
		},
	}

	for k := range uu {
		u := uu[k]
		v.SortColCmd(u.sortCol, true)(nil)
		assert.Equal(t, len(u.sorted)+1, v.GetRowCount())
		for i, s := range u.sorted {
			assert.Equal(t, s, v.GetCell(i+1, 0).Text)
		}
		v.SortInvertCmd(nil)
		assert.Equal(t, len(u.reversed)+1, v.GetRowCount())
		for i, s := range u.reversed {
			assert.Equal(t, s, v.GetCell(i+1, 0).Text)
		}
	}
}

// ----------------------------------------------------------------------------
// Helpers...

type mockTableModel struct{}

var _ ui.Tabular = (*mockTableModel)(nil)

func (t *mockTableModel) SetInstance(string)                 {}
func (t *mockTableModel) SetLabelFilter(string)              {}
func (t *mockTableModel) GetLabelFilter() string             { return "" }
func (t *mockTableModel) Empty() bool                        { return false }
func (t *mockTableModel) RowCount() int                      { return 1 }
func (t *mockTableModel) HasMetrics() bool                   { return true }
func (t *mockTableModel) Peek() *model1.TableData            { return makeTableData() }
func (t *mockTableModel) Refresh(context.Context) error      { return nil }
func (t *mockTableModel) ClusterWide() bool                  { return false }
func (t *mockTableModel) GetNamespace() string               { return "blee" }
func (t *mockTableModel) SetNamespace(string)                {}
func (t *mockTableModel) ToggleToast()                       {}
func (t *mockTableModel) AddListener(model.TableListener)    {}
func (t *mockTableModel) RemoveListener(model.TableListener) {}
func (t *mockTableModel) Watch(context.Context) error        { return nil }
func (t *mockTableModel) Get(context.Context, string) (runtime.Object, error) {
	return nil, nil
}

func (t *mockTableModel) Delete(context.Context, string, *metav1.DeletionPropagation, dao.Grace) error {
	return nil
}

func (t *mockTableModel) Describe(context.Context, string) (string, error) {
	return "", nil
}

func (t *mockTableModel) ToYAML(ctx context.Context, path string) (string, error) {
	return "", nil
}

func (t *mockTableModel) InNamespace(string) bool      { return true }
func (t *mockTableModel) SetRefreshRate(time.Duration) {}

func makeTableData() *model1.TableData {
	return model1.NewTableDataWithRows(
		client.NewGVR("test"),
		model1.Header{
			model1.HeaderColumn{Name: "NAMESPACE"},
			model1.HeaderColumn{Name: "NAME", Align: tview.AlignRight},
			model1.HeaderColumn{Name: "FRED"},
			model1.HeaderColumn{Name: "AGE", Time: true},
		},
		model1.NewRowEventsWithEvts(
			model1.RowEvent{
				Row: model1.Row{
					Fields: model1.Fields{"ns1", "r3", "10", "3y125d"},
				},
			},
			model1.RowEvent{
				Row: model1.Row{
					Fields: model1.Fields{"ns1", "r2", "15", "2y12d"},
				},
				Deltas: model1.DeltaRow{"", "", "20", ""},
			},
			model1.RowEvent{
				Row: model1.Row{
					Fields: model1.Fields{"ns1", "r1", "20", "19h"},
				},
			},
			model1.RowEvent{
				Row: model1.Row{
					Fields: model1.Fields{"ns1", "r0", "15", "10s"},
				},
			},
		),
	)
}

func makeContext() context.Context {
	a := NewApp(mock.NewMockConfig())
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)
	return context.WithValue(ctx, internal.KeyStyles, a.Styles)
}

func ensureDumpDir(n string) error {
	config.AppDumpsDir = n
	if _, err := os.Stat(n); errors.Is(err, fs.ErrNotExist) {
		return os.Mkdir(n, 0700)
	}
	if err := os.RemoveAll(n); err != nil {
		return err
	}
	return os.Mkdir(n, 0700)
}
