package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestTableDataDelete(t *testing.T) {
	uu := map[string]struct {
		re render.RowEvents
		kk map[string]struct{}
		e  render.RowEvents
	}{
		"ordered": {
			re: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			kk: map[string]struct{}{"A": {}, "C": {}},
			e: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
		},
		"unordered": {
			re: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				{Row: render.Row{ID: "D", Fields: render.Fields{"10", "2", "3"}}},
			},
			kk: map[string]struct{}{"C": {}, "A": {}},
			e: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
		},
	}

	var table render.TableData
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			table.RowEvents = u.re
			table.Delete(u.kk)
			assert.Equal(t, u.e, table.RowEvents)
		})
	}

}
