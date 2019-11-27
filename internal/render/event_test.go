package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
)

func TestSort(t *testing.T) {
	uu := map[string]struct {
		re  render.RowEvents
		col int
		asc bool
		e   render.RowEvents
	}{
		"col0": {
			re: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			col: 0,
			asc: true,
			e: render.RowEvents{
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			u.re.Sort("", u.col, u.asc)
			assert.Equal(t, u.e, u.re)
		})
	}
}

func TestDefaultColorer(t *testing.T) {
	uu := map[string]struct {
		k render.ResEvent
		e tcell.Color
	}{
		"add":    {render.EventAdd, render.AddColor},
		"update": {render.EventUpdate, render.ModColor},
		"delete": {render.EventDelete, render.KillColor},
		"std":    {100, render.StdColor},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, render.DefaultColorer("", u.k, render.Row{}))
		})
	}
}
