package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
)

func TestRowEventsDelete(t *testing.T) {
	uu := map[string]struct {
		re render.RowEvents
		id string
		e  render.RowEvents
	}{
		"first": {
			re: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			id: "A",
			e: render.RowEvents{
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
		},
		"middle": {
			re: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			id: "B",
			e: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
		},
		"last": {
			re: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			id: "C",
			e: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.re.Delete(u.id))
		})
	}
}

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
		"id_preserve": {
			re: render.RowEvents{
				{Row: render.Row{ID: "ns1/B", Fields: render.Fields{"B", "2", "3"}}},
				{Row: render.Row{ID: "ns1/A", Fields: render.Fields{"A", "2", "3"}}},
				{Row: render.Row{ID: "ns1/C", Fields: render.Fields{"C", "2", "3"}}},
				{Row: render.Row{ID: "ns2/B", Fields: render.Fields{"B", "2", "3"}}},
				{Row: render.Row{ID: "ns2/A", Fields: render.Fields{"A", "2", "3"}}},
				{Row: render.Row{ID: "ns2/C", Fields: render.Fields{"C", "2", "3"}}},
			},
			col: 1,
			asc: true,
			e: render.RowEvents{
				{Row: render.Row{ID: "ns1/A", Fields: render.Fields{"A", "2", "3"}}},
				{Row: render.Row{ID: "ns1/B", Fields: render.Fields{"B", "2", "3"}}},
				{Row: render.Row{ID: "ns1/C", Fields: render.Fields{"C", "2", "3"}}},
				{Row: render.Row{ID: "ns2/A", Fields: render.Fields{"A", "2", "3"}}},
				{Row: render.Row{ID: "ns2/B", Fields: render.Fields{"B", "2", "3"}}},
				{Row: render.Row{ID: "ns2/C", Fields: render.Fields{"C", "2", "3"}}},
			},
		},
	}

	for k := range uu {
		u := uu[k]
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

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, render.DefaultColorer("", render.RowEvent{}))
		})
	}
}
