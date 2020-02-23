package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestTableDataCustomize(t *testing.T) {
	uu := map[string]struct {
		t1   render.TableData
		cols []string
		wide bool
		e    render.TableData
	}{
		"same": {
			t1: render.TableData{
				Namespace: "fred",
				Header: render.Header{
					render.HeaderColumn{Name: "A"},
					render.HeaderColumn{Name: "B"},
					render.HeaderColumn{Name: "C"},
				},
				RowEvents: render.RowEvents{
					{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
					{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
					{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				},
			},
			cols: []string{"A", "B", "C"},
			e: render.TableData{
				Namespace: "fred",
				Header: render.Header{
					render.HeaderColumn{Name: "A"},
					render.HeaderColumn{Name: "B"},
					render.HeaderColumn{Name: "C"},
				},
				RowEvents: render.RowEvents{
					{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
					{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
					{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				},
			},
		},
		"wide-col": {
			t1: render.TableData{
				Namespace: "fred",
				Header: render.Header{
					render.HeaderColumn{Name: "A"},
					render.HeaderColumn{Name: "B", Wide: true},
					render.HeaderColumn{Name: "C"},
				},
				RowEvents: render.RowEvents{
					{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
					{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
					{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				},
			},
			cols: []string{"A", "B", "C"},
			e: render.TableData{
				Namespace: "fred",
				Header: render.Header{
					render.HeaderColumn{Name: "A"},
					render.HeaderColumn{Name: "B", Wide: false},
					render.HeaderColumn{Name: "C"},
				},
				RowEvents: render.RowEvents{
					{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
					{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
					{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				},
			},
		},
		"wide": {
			t1: render.TableData{
				Namespace: "fred",
				Header: render.Header{
					render.HeaderColumn{Name: "A"},
					render.HeaderColumn{Name: "B", Wide: true},
					render.HeaderColumn{Name: "C"},
				},
				RowEvents: render.RowEvents{
					{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
					{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
					{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				},
			},
			wide: true,
			cols: []string{"A", "C"},
			e: render.TableData{
				Namespace: "fred",
				Header: render.Header{
					render.HeaderColumn{Name: "A"},
					render.HeaderColumn{Name: "C"},
					render.HeaderColumn{Name: "B", Wide: true},
				},
				RowEvents: render.RowEvents{
					{Row: render.Row{ID: "A", Fields: render.Fields{"1", "3", "2"}}},
					{Row: render.Row{ID: "B", Fields: render.Fields{"0", "3", "2"}}},
					{Row: render.Row{ID: "C", Fields: render.Fields{"10", "3", "2"}}},
				},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.t1.Customize(u.cols, u.wide))
		})
	}
}

func TestTableDataDiff(t *testing.T) {
	uu := map[string]struct {
		t1 render.TableData
		t2 render.TableData
		e  bool
	}{
		"empty": {
			t1: render.TableData{
				Namespace: "fred",
				Header: render.Header{
					render.HeaderColumn{Name: "A"},
					render.HeaderColumn{Name: "B"},
					render.HeaderColumn{Name: "C"},
				},
				RowEvents: render.RowEvents{
					{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
					{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
					{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				},
			},
			e: true,
		},
		"same": {
			t1: render.TableData{
				Namespace: "fred",
				Header: render.Header{
					render.HeaderColumn{Name: "A"},
					render.HeaderColumn{Name: "B"},
					render.HeaderColumn{Name: "C"},
				},
				RowEvents: render.RowEvents{
					{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
					{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
					{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				},
			},
			t2: render.TableData{
				Namespace: "fred",
				Header: render.Header{
					render.HeaderColumn{Name: "A"},
					render.HeaderColumn{Name: "B"},
					render.HeaderColumn{Name: "C"},
				},
				RowEvents: render.RowEvents{
					{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
					{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
					{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				},
			},
		},
		"ns-diff": {
			t1: render.TableData{
				Namespace: "fred",
				Header: render.Header{
					render.HeaderColumn{Name: "A"},
					render.HeaderColumn{Name: "B"},
					render.HeaderColumn{Name: "C"},
				},
				RowEvents: render.RowEvents{
					{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
					{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
					{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				},
			},
			t2: render.TableData{
				Namespace: "blee",
				Header: render.Header{
					render.HeaderColumn{Name: "A"},
					render.HeaderColumn{Name: "B"},
					render.HeaderColumn{Name: "C"},
				},
				RowEvents: render.RowEvents{
					{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
					{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
					{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				},
			},
			e: true,
		},
		"header-diff": {
			t1: render.TableData{
				Namespace: "fred",
				Header: render.Header{
					render.HeaderColumn{Name: "A"},
					render.HeaderColumn{Name: "D"},
					render.HeaderColumn{Name: "C"},
				},
				RowEvents: render.RowEvents{
					{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
					{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
					{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				},
			},
			t2: render.TableData{
				Namespace: "fred",
				Header: render.Header{
					render.HeaderColumn{Name: "A"},
					render.HeaderColumn{Name: "B"},
					render.HeaderColumn{Name: "C"},
				},
				RowEvents: render.RowEvents{
					{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
					{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
					{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				},
			},
			e: true,
		},
		"row-diff": {
			t1: render.TableData{
				Namespace: "fred",
				Header: render.Header{
					render.HeaderColumn{Name: "A"},
					render.HeaderColumn{Name: "B"},
					render.HeaderColumn{Name: "C"},
				},
				RowEvents: render.RowEvents{
					{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
					{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
					{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				},
			},
			t2: render.TableData{
				Namespace: "fred",
				Header: render.Header{
					render.HeaderColumn{Name: "A"},
					render.HeaderColumn{Name: "B"},
					render.HeaderColumn{Name: "C"},
				},
				RowEvents: render.RowEvents{
					{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
					{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
					{Row: render.Row{ID: "C", Fields: render.Fields{"100", "2", "3"}}},
				},
			},
			e: true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.t1.Diff(u.t2))
		})
	}
}

func TestTableDataUpdate(t *testing.T) {
	uu := map[string]struct {
		re render.RowEvents
		rr render.Rows
		e  render.RowEvents
	}{
		"no-change": {
			re: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			rr: render.Rows{
				render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
				render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}},
				render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}},
			},
			e: render.RowEvents{
				{Kind: render.EventUnchanged, Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Kind: render.EventUnchanged, Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Kind: render.EventUnchanged, Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
		},
		"add": {
			re: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			rr: render.Rows{
				render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
				render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}},
				render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}},
				render.Row{ID: "D", Fields: render.Fields{"10", "2", "3"}},
			},
			e: render.RowEvents{
				{Kind: render.EventUnchanged, Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Kind: render.EventUnchanged, Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Kind: render.EventUnchanged, Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				{Kind: render.EventAdd, Row: render.Row{ID: "D", Fields: render.Fields{"10", "2", "3"}}},
			},
		},
		"delete": {
			re: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			rr: render.Rows{
				render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
				render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}},
			},
			e: render.RowEvents{
				{Kind: render.EventUnchanged, Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Kind: render.EventUnchanged, Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
		},
		"update": {
			re: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			rr: render.Rows{
				render.Row{ID: "A", Fields: render.Fields{"10", "2", "3"}},
				render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}},
				render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}},
			},
			e: render.RowEvents{
				{
					Kind:   render.EventUpdate,
					Row:    render.Row{ID: "A", Fields: render.Fields{"10", "2", "3"}},
					Deltas: render.DeltaRow{"1", "", ""},
				},
				{Kind: render.EventUnchanged, Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Kind: render.EventUnchanged, Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
		},
	}

	var table render.TableData
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			table.RowEvents = u.re
			table.Update(u.rr)
			assert.Equal(t, u.e, table.RowEvents)
		})
	}
}

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
