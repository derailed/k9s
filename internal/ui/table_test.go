package ui_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestTableNew(t *testing.T) {
	v := ui.NewTable("fred")
	s, _ := config.NewStyles("")
	ctx := context.WithValue(context.Background(), ui.KeyStyles, s)
	v.Init(ctx)

	assert.Equal(t, "fred", v.BaseTitle)
}

func TestTableUpdate(t *testing.T) {
	v := ui.NewTable("fred")
	s, _ := config.NewStyles("")
	ctx := context.WithValue(context.Background(), ui.KeyStyles, s)
	v.Init(ctx)

	v.Update(makeTableData())

	assert.Equal(t, 3, v.GetRowCount())
	assert.Equal(t, 3, v.GetColumnCount())
}

func TestTableSelection(t *testing.T) {
	v := ui.NewTable("fred")
	s, _ := config.NewStyles("")
	ctx := context.WithValue(context.Background(), ui.KeyStyles, s)
	v.Init(ctx)
	v.Update(makeTableData())
	v.SelectRow(1, true)

	assert.True(t, v.RowSelected())
	assert.Equal(t, resource.Row{"blee", "duh", "fred"}, v.GetRow())
	assert.Equal(t, "blee", v.GetSelectedCell(0))
	assert.Equal(t, 1, v.GetSelectedRowIndex())
	assert.Equal(t, []string{"r1"}, v.GetSelectedItems())

	v.ClearSelection()
	v.SelectFirstRow()
	assert.Equal(t, 1, v.GetSelectedRowIndex())
}

// Helpers...

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
