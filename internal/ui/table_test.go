package ui_test

import (
	"context"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestTableNew(t *testing.T) {
	v := ui.NewTable("fred")
	ctx := context.WithValue(context.Background(), ui.KeyStyles, config.NewStyles())
	v.Init(ctx)

	assert.Equal(t, "fred", v.BaseTitle)
}

func TestTableUpdate(t *testing.T) {
	v := ui.NewTable("fred")
	ctx := context.WithValue(context.Background(), ui.KeyStyles, config.NewStyles())
	v.Init(ctx)

	v.Update(makeTableData())

	assert.Equal(t, 3, v.GetRowCount())
	assert.Equal(t, 3, v.GetColumnCount())
}

func TestTableSelection(t *testing.T) {
	v := ui.NewTable("fred")
	ctx := context.WithValue(context.Background(), ui.KeyStyles, config.NewStyles())
	v.Init(ctx)
	m := &testModel{}
	v.SetModel(m)
	v.Update(m.Peek())
	v.SelectRow(1, true)

	assert.Equal(t, "r1", v.GetSelectedItem())
	assert.Equal(t, render.Row{ID: "r2", Fields: render.Fields{"blee", "duh", "zorg"}}, v.GetSelectedRow())
	assert.Equal(t, "blee", v.GetSelectedCell(0))
	assert.Equal(t, 1, v.GetSelectedRowIndex())
	assert.Equal(t, []string{"r1"}, v.GetSelectedItems())

	v.ClearSelection()
	v.SelectFirstRow()
	assert.Equal(t, 1, v.GetSelectedRowIndex())
}

// ----------------------------------------------------------------------------
// Helpers...

type testModel struct{}

var _ ui.Tabular = &testModel{}

func (t *testModel) Empty() bool                     { return false }
func (t *testModel) Peek() render.TableData          { return makeTableData() }
func (t *testModel) ClusterWide() bool               { return false }
func (t *testModel) GetNamespace() string            { return "blee" }
func (t *testModel) SetNamespace(string)             {}
func (t *testModel) AddListener(model.TableListener) {}
func (t *testModel) Watch(context.Context)           {}
func (t *testModel) InNamespace(string) bool         { return true }
func (t *testModel) SetRefreshRate(time.Duration)    {}

func makeTableData() render.TableData {
	return render.TableData{
		Namespace: "",
		Header: render.HeaderRow{
			render.Header{Name: "a"},
			render.Header{Name: "b"},
			render.Header{Name: "c"},
		},
		RowEvents: render.RowEvents{
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
		},
	}
}
