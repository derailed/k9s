package view

import (
	"context"
	"os"
	"path/filepath"
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
	v.SetModel(&mockTableModel{})
	v.Refresh()
	v.CmdBuff().SetActive(true)
	v.CmdBuff().SetText("blee", "")

	assert.Equal(t, 3, v.GetRowCount())
}

func TestTableViewSort(t *testing.T) {
	v := NewTable(client.NewGVR("test"))
	v.Init(makeContext())
	v.SetModel(&mockTableModel{})
	v.SortColCmd("NAME", true)(nil)
	assert.Equal(t, 3, v.GetRowCount())
	assert.Equal(t, "blee", v.GetCell(1, 0).Text)

	v.SortInvertCmd(nil)
	assert.Equal(t, 3, v.GetRowCount())
	assert.Equal(t, "fred", v.GetCell(1, 0).Text)
}

// ----------------------------------------------------------------------------
// Helpers...

type mockTableModel struct{}

var _ ui.Tabular = (*mockTableModel)(nil)

func (t *mockTableModel) SetInstance(string)                 {}
func (t *mockTableModel) SetLabelFilter(string)              {}
func (t *mockTableModel) Empty() bool                        { return false }
func (t *mockTableModel) HasMetrics() bool                   { return true }
func (t *mockTableModel) Peek() render.TableData             { return makeTableData() }
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

func (t *mockTableModel) Delete(context.Context, string, bool, bool) error {
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
