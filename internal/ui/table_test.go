package ui_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/watch"
)

func TestTableNew(t *testing.T) {
	v := ui.NewTable("fred")
	s, _ := config.NewStyles("")
	ctx := context.WithValue(context.Background(), ui.KeyStyles, s)
	v.Init(ctx)

	assert.Equal(t, "fred", v.GetBaseTitle())

	v.SetBaseTitle("bozo")
	assert.Equal(t, "bozo", v.GetBaseTitle())

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
	assert.Equal(t, []string{"blee/duh"}, v.GetSelectedItems())

	v.ClearSelection()
	v.SelectFirstRow()
	assert.Equal(t, 1, v.GetSelectedRowIndex())
}

// Helpers...

func makeTableData() resource.TableData {
	return resource.TableData{
		Namespace: "",
		Header:    resource.Row{"a", "b", "c"},
		Rows: resource.RowEvents{
			"r1": &resource.RowEvent{
				Action: watch.Added,
				Fields: resource.Row{"blee", "duh", "fred"},
				Deltas: resource.Row{"", "", ""},
			},
			"r2": &resource.RowEvent{
				Action: watch.Added,
				Fields: resource.Row{"fred", "duh", "zorg"},
				Deltas: resource.Row{"", "", ""},
			},
		},
	}
}
