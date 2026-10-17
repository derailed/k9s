// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"strings"
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
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestHeaderIndex(t *testing.T) {
	uu := map[string]struct {
		colName string
		cells   []string
		eok     bool
		e       int
	}{
		"simple": {
			cells:   []string{"NAMESPACE", "NAME", "AGE"},
			colName: "NAME",
			eok:     true,
			e:       1,
		},

		"missing": {
			cells:   []string{"NAMESPACE", "BLEE", "AGE"},
			colName: "NAME",
		},

		"decorated": {
			cells:   []string{"[#DADEE8::]NAMESPACE[::]", "[#DADEE8::]NAME[::]", "[#DADEE8::]AGE[::]"},
			colName: "NAME",
			eok:     true,
			e:       1,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			ta := NewTable(client.NewGVR("test"))
			require.NoError(t, ta.Init(makeContext(t)))
			for i, c := range u.cells {
				ta.AddHeaderCell(i, model1.HeaderColumn{Name: c})
			}

			i, ok := ta.HeaderIndex(u.colName)
			assert.Equal(t, u.eok, ok)
			assert.Equal(t, u.e, i)
		})
	}
}

func TestTableSave(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	require.NoError(t, v.Init(makeContext(t)))
	v.SetTitle("k9s-test")

	require.NoError(t, ensureDumpDir("/tmp/test-dumps"))
	dir := v.app.Config.K9s.ContextScreenDumpDir()
	c1, _ := os.ReadDir(dir)
	v.saveCmd(nil)

	c2, _ := os.ReadDir(dir)
	assert.Len(t, c2, len(c1)+1)
}

func TestTableNew(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	require.NoError(t, v.Init(makeContext(t)))

	data := model1.NewTableDataWithRows(
		client.NewGVR("test"),
		model1.Header{
			model1.HeaderColumn{Name: "NAMESPACE"},
			model1.HeaderColumn{Name: "NAME", Attrs: model1.Attrs{Align: tview.AlignRight}},
			model1.HeaderColumn{Name: "FRED"},
			model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true, Decorator: render.AgeDecorator}},
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
	require.NoError(t, v.Init(makeContext(t)))
	v.SetModel(&mockTableModel{})
	v.Refresh()

	v.CmdBuff().SetActive(true)
	v.CmdBuff().SetText("blee", "", true)

	assert.Equal(t, 5, v.GetRowCount())
}

func TestTableViewSort(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	require.NoError(t, v.Init(makeContext(t)))
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
		assert.Len(t, u.sorted, v.GetRowCount()-1)
		for i, s := range u.sorted {
			assert.Equal(t, s, v.GetCell(i+1, 0).Text)
		}
		v.SortInvertCmd(nil)
		assert.Len(t, u.reversed, v.GetRowCount()-1)
		for i, s := range u.reversed {
			assert.Equal(t, s, v.GetCell(i+1, 0).Text)
		}
	}
}

func TestTableSortCapturesHistory(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	require.NoError(t, v.Init(makeContext(t)))
	v.SetModel(new(mockTableModel))
	v.SetCommand(cmd.NewInterpreter("test"))
	v.app.cmdHistory.Push("test")

	v.SortColCmd("NAME", true)(nil)

	top, ok := v.app.cmdHistory.Top()
	require.True(t, ok)
	assert.Equal(t, "test sort:name:asc", top)
}

func TestTableRestoreSort(t *testing.T) {
	v := NewTable(client.NewGVR("test"))

	// Restore in Command.run happens BEFORE the component is initialized.
	v.SetSortCol("FRED", true)
	v.SetManualSort(true)

	require.NoError(t, v.Init(makeContext(t)))
	v.SetCommand(cmd.NewInterpreter("test sort:fred:asc"))
	// Start registers the table as a view-config listener, which fires
	// ViewSettingsChanged and may reset the restored manual sort.
	v.Start()

	data := model1.NewTableDataWithRows(
		client.NewGVR("test"),
		model1.Header{
			model1.HeaderColumn{Name: "NAMESPACE"},
			model1.HeaderColumn{Name: "NAME"},
			model1.HeaderColumn{Name: "FRED"},
			model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true, Decorator: render.AgeDecorator}},
		},
		model1.NewRowEventsWithEvts(
			model1.RowEvent{Row: model1.Row{Fields: model1.Fields{"ns1", "a", "30", "3m"}}},
			model1.RowEvent{Row: model1.Row{Fields: model1.Fields{"ns1", "b", "10", "1m"}}},
			model1.RowEvent{Row: model1.Row{Fields: model1.Fields{"ns1", "c", "20", "2m"}}},
		),
	)
	cdata := v.Update(data, false)
	v.UpdateUI(cdata, data)

	assert.Equal(t, "FRED", v.GetSortCol().Name)
	// FRED ascending => 10,20,30 => rows b,c,a
	assert.Equal(t, "b", strings.TrimSpace(v.GetCell(1, 1).Text))
	assert.Equal(t, "c", strings.TrimSpace(v.GetCell(2, 1).Text))
	assert.Equal(t, "a", strings.TrimSpace(v.GetCell(3, 1).Text))
	// Visible cols: NAMESPACE=0, NAME=1, FRED=2.
	// The highlighted column must follow the restored sort column.
	assert.Equal(t, 2, v.GetSelectedColIdx())
}

func TestTableRestoreSortWithCustomView(t *testing.T) {
	v := NewTable(client.NewGVR("test"))

	// Restore in Command.run happens BEFORE the component is initialized.
	v.SetSortCol("FRED", true)
	v.SetManualSort(true)

	require.NoError(t, v.Init(makeContext(t)))
	v.SetCommand(cmd.NewInterpreter("test sort:fred:asc"))

	// Simulate a custom view config (views.yaml) with a different default
	// sort column. Registering the listener fires ViewSettingsChanged which
	// must not clobber the restored sort.
	v.app.CustomView().Views["test"] = config.ViewSetting{SortColumn: "NAME:desc"}
	v.Start()

	data := model1.NewTableDataWithRows(
		client.NewGVR("test"),
		model1.Header{
			model1.HeaderColumn{Name: "NAMESPACE"},
			model1.HeaderColumn{Name: "NAME"},
			model1.HeaderColumn{Name: "FRED"},
			model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true, Decorator: render.AgeDecorator}},
		},
		model1.NewRowEventsWithEvts(
			model1.RowEvent{Row: model1.Row{Fields: model1.Fields{"ns1", "a", "30", "3m"}}},
			model1.RowEvent{Row: model1.Row{Fields: model1.Fields{"ns1", "b", "10", "1m"}}},
			model1.RowEvent{Row: model1.Row{Fields: model1.Fields{"ns1", "c", "20", "2m"}}},
		),
	)
	cdata := v.Update(data, false)
	v.UpdateUI(cdata, data)

	assert.Equal(t, "FRED", v.GetSortCol().Name)
	// FRED ascending => 10,20,30 => rows b,c,a
	assert.Equal(t, "b", strings.TrimSpace(v.GetCell(1, 1).Text))
	assert.Equal(t, "c", strings.TrimSpace(v.GetCell(2, 1).Text))
	assert.Equal(t, "a", strings.TrimSpace(v.GetCell(3, 1).Text))
	// Visible cols: NAMESPACE=0, NAME=1, FRED=2.
	// The highlighted column must follow the restored sort column.
	assert.Equal(t, 2, v.GetSelectedColIdx())
}

// ----------------------------------------------------------------------------
// Helpers...

type mockTableModel struct{}

var _ ui.Tabular = (*mockTableModel)(nil)

func (*mockTableModel) SetViewSetting(context.Context, *config.ViewSetting) {}
func (*mockTableModel) SetInstance(string)                                  {}
func (*mockTableModel) SetLabelSelector(labels.Selector)                    {}
func (*mockTableModel) GetLabelSelector() labels.Selector                   { return nil }
func (*mockTableModel) Empty() bool                                         { return false }
func (*mockTableModel) RowCount() int                                       { return 1 }
func (*mockTableModel) HasMetrics() bool                                    { return true }
func (*mockTableModel) Peek() *model1.TableData                             { return makeTableData() }
func (*mockTableModel) Refresh(context.Context) error                       { return nil }
func (*mockTableModel) ClusterWide() bool                                   { return false }
func (*mockTableModel) GetNamespace() string                                { return "blee" }
func (*mockTableModel) SetNamespace(string)                                 {}
func (*mockTableModel) ToggleToast()                                        {}
func (*mockTableModel) AddListener(model.TableListener)                     {}
func (*mockTableModel) RemoveListener(model.TableListener)                  {}
func (*mockTableModel) Watch(context.Context) error                         { return nil }
func (*mockTableModel) Get(context.Context, string) (runtime.Object, error) {
	return nil, nil
}
func (*mockTableModel) Delete(context.Context, string, *metav1.DeletionPropagation, dao.Grace) error {
	return nil
}
func (*mockTableModel) Describe(context.Context, string) (string, error) {
	return "", nil
}
func (*mockTableModel) ToYAML(context.Context, string) (string, error) {
	return "", nil
}
func (*mockTableModel) InNamespace(string) bool      { return true }
func (*mockTableModel) SetRefreshRate(time.Duration) {}

func makeTableData() *model1.TableData {
	return model1.NewTableDataWithRows(
		client.NewGVR("test"),
		model1.Header{
			model1.HeaderColumn{Name: "NAMESPACE"},
			model1.HeaderColumn{Name: "NAME", Attrs: model1.Attrs{Align: tview.AlignRight}},
			model1.HeaderColumn{Name: "FRED"},
			model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
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

func makeContext(t *testing.T) context.Context {
	a := NewApp(mock.NewMockConfig(t))
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
