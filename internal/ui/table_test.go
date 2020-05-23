package ui_test

import (
	"context"
	"testing"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
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
	v.Update(data)

	assert.Equal(t, len(data.RowEvents)+1, v.GetRowCount())
	assert.Equal(t, len(data.Header), v.GetColumnCount())
}

func TestTableSelection(t *testing.T) {
	v := ui.NewTable(client.NewGVR("fred"))
	v.Init(makeContext())
	m := &mockModel{}
	v.SetModel(m)
	v.Update(m.Peek())
	v.SelectRow(1, true)

	r, ok := v.GetSelectedRow("r1")
	assert.True(t, ok)
	assert.Equal(t, "r1", v.GetSelectedItem())
	assert.Equal(t, render.Row{ID: "r1", Fields: render.Fields{"blee", "duh", "fred"}}, r)
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

func (t *mockModel) SetInstance(string)                 {}
func (t *mockModel) Empty() bool                        { return false }
func (t *mockModel) HasMetrics() bool                   { return true }
func (t *mockModel) Peek() render.TableData             { return makeTableData() }
func (t *mockModel) Refresh(context.Context)            {}
func (t *mockModel) ClusterWide() bool                  { return false }
func (t *mockModel) GetNamespace() string               { return "blee" }
func (t *mockModel) SetNamespace(string)                {}
func (t *mockModel) ToggleToast()                       {}
func (t *mockModel) AddListener(model.TableListener)    {}
func (t *mockModel) RemoveListener(model.TableListener) {}
func (t *mockModel) Watch(context.Context)              {}
func (t *mockModel) Get(ctx context.Context, path string) (runtime.Object, error) {
	return nil, nil
}
func (t *mockModel) Delete(ctx context.Context, path string, c, f bool) error {
	return nil
}
func (t *mockModel) Describe(context.Context, string) (string, error) {
	return "", nil
}
func (t *mockModel) ToYAML(ctx context.Context, path string) (string, error) {
	return "", nil
}
func (t *mockModel) InNamespace(string) bool      { return true }
func (t *mockModel) SetRefreshRate(time.Duration) {}

func makeTableData() render.TableData {
	t := render.NewTableData()
	t.Namespace = ""
	t.Header = render.Header{
		render.HeaderColumn{Name: "A"},
		render.HeaderColumn{Name: "B"},
		render.HeaderColumn{Name: "C"},
	}
	t.RowEvents = render.RowEvents{
		render.RowEvent{
			Row: render.Row{
				ID:     "r1",
				Fields: render.Fields{"blee", "duh", "fred"},
			},
		},
		render.RowEvent{
			Row: render.Row{
				ID:     "r2",
				Fields: render.Fields{"blee", "duh", "zorg"},
			},
		},
	}

	return *t
}

func makeContext() context.Context {
	ctx := context.WithValue(context.Background(), internal.KeyStyles, config.NewStyles())
	ctx = context.WithValue(ctx, internal.KeyViewConfig, config.NewCustomView())

	return ctx
}
