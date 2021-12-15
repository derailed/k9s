package view

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestTableSave(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	v.Init(makeContext())
	v.SetTitle("k9s-test")

	dir := filepath.Join(v.app.Config.K9s.GetScreenDumpDir(), v.app.Config.K9s.CurrentCluster)
	c1, _ := os.ReadDir(dir)
	v.saveCmd(nil)

	c2, _ := os.ReadDir(dir)
	assert.Equal(t, len(c2), len(c1)+1)
}

func TestTableNew(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	v.Init(makeContext())

	data := render.NewTableData()
	data.Header = render.Header{
		render.HeaderColumn{Name: "NAMESPACE"},
		render.HeaderColumn{Name: "NAME", Align: tview.AlignRight},
		render.HeaderColumn{Name: "FRED"},
		render.HeaderColumn{Name: "AGE", Time: true, Decorator: render.AgeDecorator},
	}
	data.RowEvents = render.RowEvents{
		render.RowEvent{
			Row: render.Row{
				Fields: render.Fields{"ns1", "a", "10", "3m"},
			},
		},
		render.RowEvent{
			Row: render.Row{
				Fields: render.Fields{"ns1", "b", "15", "1m"},
			},
		},
	}
	data.Namespace = ""

	v.Update(*data, false)
	assert.Equal(t, 3, v.GetRowCount())
}

func TestTableViewFilter(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	v.Init(makeContext())
	v.SetModel(defaultModelMock())
	v.Refresh()
	v.CmdBuff().SetActive(true)
	v.CmdBuff().SetText("blee", "")

	assert.Equal(t, 3, v.GetRowCount())
}

func TestTableViewSort(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	v.Init(makeContext())
	v.SetModel(defaultModelMock())
	v.SortColCmd("NAME", true)(nil)
	assert.Equal(t, 3, v.GetRowCount())
	assert.Equal(t, "blee", v.GetCell(1, 0).Text)

	v.SortInvertCmd(nil)
	assert.Equal(t, 3, v.GetRowCount())
	assert.Equal(t, "fred", v.GetCell(1, 0).Text)
}

func TestTableViewSortChangeSmoke(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	v.Init(makeContext())
	v.SetModel(defaultModelMock())
	v.SortColCmd("NAME", true)(nil)
	assert.Equal(t, 3, v.GetRowCount())
	assert.Equal(t, "blee", v.GetCell(1, 0).Text)

	v.SortColChange(ui.SortNextCol)(nil)
	v.SortColChange(ui.SortNextCol)(nil)
	sortCol, asc := v.GetSortCol()
	assert.Equal(t, "AGE", sortCol)
	assert.Equal(t, true, asc)
	assert.Equal(t, "fred", v.GetCell(1, 0).Text)

	v.SortInvertCmd(nil)
	_, asc = v.GetSortCol()
	assert.Equal(t, false, asc)
	assert.Equal(t, "blee", v.GetCell(1, 0).Text)
}

type SortSteps []struct {
	change  ui.SortChange
	reverse bool

	sortCol string
	asc     bool
	value   string
	col     int
}

func TestTableViewSortWrapAround(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	v.Init(makeContext())
	v.SetModel(modelMockForSortingMinimal())
	v.SortColCmd("NAME", true)(nil)
	assert.Equal(t, 3, v.GetRowCount())

	steps := SortSteps{
		{change: ui.SortNextCol, sortCol: "FRED", asc: true, value: "fred1", col: 1},
		{change: ui.SortNextCol, sortCol: "AGE", asc: true, value: "90s", col: 2},
		{change: ui.SortNextCol, sortCol: "NAME", asc: true, value: "name1", col: 0},
		{change: ui.SortNextCol, reverse: true, sortCol: "FRED", asc: false, value: "fred2", col: 1},
		{change: ui.SortPrevCol, sortCol: "NAME", asc: false, value: "name2", col: 0},
		{change: ui.SortPrevCol, sortCol: "AGE", asc: false, value: "110s", col: 2},
		{change: ui.SortPrevCol, sortCol: "FRED", asc: false, value: "fred2", col: 1},
	}

	runSortSteps(t, v, steps)
}

func TestTableViewFullSortWrapAround(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	v.ToggleWide() // Enable wide display
	v.Init(makeContext())
	v.SetModel(modelMockForSortingFull())
	v.Update(makeTableDataForSorting(), true) // Enable metrics
	v.SortColCmd("NAME", true)(nil)
	assert.Equal(t, 3, v.GetRowCount())

	steps := SortSteps{
		{change: ui.SortNextCol, sortCol: "LABELS", asc: true, value: "k8s-app=kube-dns1", col: 2},
		{change: ui.SortNextCol, sortCol: "FRED", asc: true, value: "fred1", col: 3},
		{change: ui.SortNextCol, sortCol: "CPU", asc: true, value: "10", col: 4},
		{change: ui.SortNextCol, sortCol: "AGE", asc: true, value: "90s", col: 5},
		{change: ui.SortNextCol, sortCol: "NAMESPACE", asc: true, value: "ns1", col: 0},
		{change: ui.SortNextCol, reverse: true, sortCol: "NAME", asc: false, value: "name2", col: 1},
		{change: ui.SortPrevCol, sortCol: "NAMESPACE", asc: false, value: "ns1", col: 0},
		{change: ui.SortPrevCol, sortCol: "AGE", asc: false, value: "110s", col: 5},
		{change: ui.SortPrevCol, sortCol: "CPU", asc: false, value: "20", col: 4},
		{change: ui.SortPrevCol, sortCol: "FRED", asc: false, value: "fred2", col: 3},
		{change: ui.SortPrevCol, sortCol: "LABELS", asc: false, value: "k8s-app=kube-dns2", col: 2},
	}

	runSortSteps(t, v, steps)
}

func runSortSteps(t *testing.T, v *Table, steps SortSteps) {
	for _, step := range steps {
		v.SortColChange(step.change)(nil)
		if step.reverse {
			v.SortInvertCmd(nil)
		}
		sortCol, asc := v.GetSortCol()
		assert.Equal(t, step.sortCol, sortCol)
		assert.Equal(t, step.asc, asc)
		assert.Equal(t, step.value, strings.TrimSpace(v.GetCell(1, step.col).Text))
	}
}

func TestTableViewSortCapacity(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	v.Init(makeContext())

	data := render.NewTableData()
	data.Header = render.Header{
		render.HeaderColumn{Name: "NAMESPACE"},
		render.HeaderColumn{Name: "NAME", Align: tview.AlignRight},
		render.HeaderColumn{Name: "CAPACITY"},
		render.HeaderColumn{Name: "AGE", Time: true, Decorator: render.AgeDecorator},
	}
	data.RowEvents = render.RowEvents{
		render.RowEvent{
			Row: render.Row{
				Fields: render.Fields{"ns1", "a", "100Mi", "3m"},
			},
		},
		render.RowEvent{
			Row: render.Row{
				Fields: render.Fields{"ns1", "b", "900Mi", "1m"},
			},
		},
		render.RowEvent{
			Row: render.Row{
				Fields: render.Fields{"ns1", "c", "8Gi", "10m"},
			},
		},
	}
	data.Namespace = ""

	v.SetSortCol("CAPACITY", true)
	v.Update(*data, false)
	assert.Equal(t, "100Mi", strings.TrimSpace(v.GetCell(1, 2).Text))
	assert.Equal(t, "900Mi", strings.TrimSpace(v.GetCell(2, 2).Text))
	assert.Equal(t, "8Gi", strings.TrimSpace(v.GetCell(3, 2).Text))
	v.SetSortCol("CAPACITY", false)
	v.Update(*data, false)
	assert.Equal(t, "8Gi", strings.TrimSpace(v.GetCell(1, 2).Text))
	assert.Equal(t, "900Mi", strings.TrimSpace(v.GetCell(2, 2).Text))
	assert.Equal(t, "100Mi", strings.TrimSpace(v.GetCell(3, 2).Text))
}

// ----------------------------------------------------------------------------
// Helpers...

type tableModelMock struct {
	MockHasMetrics  func() bool
	MockPeek        func() render.TableData
	MockClusterWide func() bool
}

func (fake *tableModelMock) HasMetrics() bool                                        { return fake.MockHasMetrics() }
func (fake *tableModelMock) Peek() render.TableData                                  { return fake.MockPeek() }
func (fake *tableModelMock) ClusterWide() bool                                       { return fake.MockClusterWide() }
func (fake *tableModelMock) SetInstance(string)                                      {}
func (fake *tableModelMock) SetLabelFilter(string)                                   {}
func (fake *tableModelMock) Empty() bool                                             { return false }
func (fake *tableModelMock) Refresh(context.Context) error                           { return nil }
func (fake *tableModelMock) GetNamespace() string                                    { return "blee" }
func (fake *tableModelMock) SetNamespace(string)                                     {}
func (fake *tableModelMock) ToggleToast()                                            {}
func (fake *tableModelMock) AddListener(model.TableListener)                         {}
func (fake *tableModelMock) RemoveListener(model.TableListener)                      {}
func (fake *tableModelMock) Watch(context.Context) error                             { return nil }
func (fake *tableModelMock) Get(context.Context, string) (runtime.Object, error)     { return nil, nil }
func (fake *tableModelMock) Delete(context.Context, string, bool, bool) error        { return nil }
func (fake *tableModelMock) Describe(context.Context, string) (string, error)        { return "", nil }
func (fake *tableModelMock) ToYAML(ctx context.Context, path string) (string, error) { return "", nil }
func (fake *tableModelMock) InNamespace(string) bool                                 { return true }
func (fake *tableModelMock) SetRefreshRate(time.Duration)                            {}

func defaultModelMock() *tableModelMock {
	return &tableModelMock{
		MockHasMetrics:  func() bool { return true },
		MockPeek:        func() render.TableData { return makeTableData() },
		MockClusterWide: func() bool { return false },
	}
}

func modelMockForSortingMinimal() *tableModelMock {
	return &tableModelMock{
		MockHasMetrics:  func() bool { return false },
		MockPeek:        func() render.TableData { return makeTableDataForSorting() },
		MockClusterWide: func() bool { return false },
	}
}

func modelMockForSortingFull() *tableModelMock {
	return &tableModelMock{
		MockHasMetrics:  func() bool { return true },
		MockPeek:        func() render.TableData { return makeTableDataForSorting() },
		MockClusterWide: func() bool { return true },
	}
}

func makeTableData() render.TableData {
	t := render.NewTableData()

	t.Header = render.Header{
		render.HeaderColumn{Name: "NAMESPACE"},
		render.HeaderColumn{Name: "NAME", Align: tview.AlignRight},
		render.HeaderColumn{Name: "FRED"},
		render.HeaderColumn{Name: "AGE", Time: true, Decorator: render.AgeDecorator},
	}
	t.RowEvents = render.RowEvents{
		render.RowEvent{
			Row: render.Row{
				Fields: render.Fields{"ns1", "blee", "10", "3m"},
			},
		},
		render.RowEvent{
			Row: render.Row{
				Fields: render.Fields{"ns1", "fred", "15", "1m"},
			},
			Deltas: render.DeltaRow{"", "", "20", ""},
		},
	}
	t.Namespace = ""

	return *t
}

func makeTableDataForSorting() render.TableData {
	t := render.NewTableData()

	t.Header = render.Header{
		render.HeaderColumn{Name: "NAMESPACE"},
		render.HeaderColumn{Name: "NAME", Align: tview.AlignRight},
		render.HeaderColumn{Name: "LABELS", Wide: true},
		render.HeaderColumn{Name: "FRED"},
		render.HeaderColumn{Name: "CPU", MX: true},
		render.HeaderColumn{Name: "AGE", Time: true, Decorator: render.AgeDecorator},
	}
	t.RowEvents = render.RowEvents{
		render.RowEvent{
			Row: render.Row{
				Fields: render.Fields{"ns1", "name1", "k8s-app=kube-dns2", "fred2", "20", "90s"},
			},
		},
		render.RowEvent{
			Row: render.Row{
				Fields: render.Fields{"ns1", "name2", "k8s-app=kube-dns1", "fred1", "10", "110s"},
			},
		},
	}
	t.Namespace = ""

	return *t
}

func makeContext() context.Context {
	a := NewApp(config.NewConfig(ks{}))
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)
	return context.WithValue(ctx, internal.KeyStyles, a.Styles)
}

type ks struct{}

func (k ks) CurrentContextName() (string, error) {
	return "test", nil
}

func (k ks) CurrentClusterName() (string, error) {
	return "test", nil
}

func (k ks) CurrentNamespaceName() (string, error) {
	return "test", nil
}

func (k ks) ClusterNames() (map[string]struct{}, error) {
	return map[string]struct{}{"test": {}}, nil
}

func (k ks) NamespaceNames(nn []v1.Namespace) []string {
	return []string{"test"}
}
