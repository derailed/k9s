// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestTableDataCustomize(t *testing.T) {
	uu := map[string]struct {
		t1, e        *TableData
		vs           config.ViewSetting
		sc           SortColumn
		wide, manual bool
	}{
		"same": {
			t1: NewTableDataWithRows(
				client.NewGVR("test"),
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "B"},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				),
			),
			vs: config.ViewSetting{Columns: []string{"A", "B", "C"}},
			e: NewTableDataWithRows(
				client.NewGVR("test"),
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "B"},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				),
			),
		},
		"wide-col": {
			t1: NewTableDataWithRows(
				client.NewGVR("test"),
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "B", Wide: true},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				),
			),
			vs: config.ViewSetting{Columns: []string{"A", "B", "C"}},
			e: NewTableDataWithRows(
				client.NewGVR("test"),
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "B", Wide: false},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				),
			),
		},
		"wide": {
			t1: NewTableDataWithRows(
				client.NewGVR("test"),
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "B", Wide: true},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				),
			),
			wide: true,
			vs:   config.ViewSetting{Columns: []string{"A", "C"}},
			e: NewTableDataWithRows(
				client.NewGVR("test"),
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "C"},
					HeaderColumn{Name: "B", Wide: true},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "3", "2"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "3", "2"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "3", "2"}}},
				),
			),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			td, _ := u.t1.Customize(&u.vs, u.sc, u.manual, u.wide)
			assert.Equal(t, u.e, td)
		})
	}
}

func TestTableDataDiff(t *testing.T) {
	uu := map[string]struct {
		t1, t2 *TableData
		e      bool
	}{
		"empty": {
			t1: NewTableDataWithRows(
				client.NewGVR("test"),
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "B"},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				),
			),
			e: true,
		},
		"same": {
			t1: NewTableDataWithRows(
				client.NewGVR("test"),
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "B"},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				),
			),
			t2: NewTableDataWithRows(
				client.NewGVR("test"),
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "B"},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				),
			),
		},
		"ns-diff": {
			t1: NewTableDataFull(
				client.NewGVR("test"),
				"ns1",
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "B"},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				),
			),
			t2: NewTableDataFull(
				client.NewGVR("test"),
				"ns-2",
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "B"},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				),
			),
			e: true,
		},
		"header-diff": {
			t1: NewTableDataWithRows(
				client.NewGVR("test"),
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "D"},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				),
			),
			t2: NewTableDataWithRows(
				client.NewGVR("test"),
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "B"},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				),
			),
			e: true,
		},
		"row-diff": {
			t1: NewTableDataWithRows(
				client.NewGVR("test"),
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "B"},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				),
			),
			t2: NewTableDataWithRows(
				client.NewGVR("test"),
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "B"},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"100", "2", "3"}}},
				),
			),
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
		re, e *RowEvents
		rr    Rows
	}{
		"no-change": {
			re: NewRowEventsWithEvts(
				RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
				RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
				RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
			),
			rr: Rows{
				Row{ID: "A", Fields: Fields{"1", "2", "3"}},
				Row{ID: "B", Fields: Fields{"0", "2", "3"}},
				Row{ID: "C", Fields: Fields{"10", "2", "3"}},
			},
			e: NewRowEventsWithEvts(
				RowEvent{Kind: EventUnchanged, Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
				RowEvent{Kind: EventUnchanged, Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
				RowEvent{Kind: EventUnchanged, Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
			),
		},
		"add": {
			re: NewRowEventsWithEvts(
				RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
				RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
				RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
			),
			rr: Rows{
				Row{ID: "A", Fields: Fields{"1", "2", "3"}},
				Row{ID: "B", Fields: Fields{"0", "2", "3"}},
				Row{ID: "C", Fields: Fields{"10", "2", "3"}},
				Row{ID: "D", Fields: Fields{"10", "2", "3"}},
			},
			e: NewRowEventsWithEvts(
				RowEvent{Kind: EventUnchanged, Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
				RowEvent{Kind: EventUnchanged, Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
				RowEvent{Kind: EventUnchanged, Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				RowEvent{Kind: EventAdd, Row: Row{ID: "D", Fields: Fields{"10", "2", "3"}}},
			),
		},
		"delete": {
			re: NewRowEventsWithEvts(
				RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
				RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
				RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
			),
			rr: Rows{
				Row{ID: "A", Fields: Fields{"1", "2", "3"}},
				Row{ID: "C", Fields: Fields{"10", "2", "3"}},
			},
			e: NewRowEventsWithEvts(
				RowEvent{Kind: EventUnchanged, Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
				RowEvent{Kind: EventUnchanged, Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
			),
		},
		"update": {
			re: NewRowEventsWithEvts(
				RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
				RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
				RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
			),
			rr: Rows{
				Row{ID: "A", Fields: Fields{"10", "2", "3"}},
				Row{ID: "B", Fields: Fields{"0", "2", "3"}},
				Row{ID: "C", Fields: Fields{"10", "2", "3"}},
			},
			e: NewRowEventsWithEvts(
				RowEvent{
					Kind:   EventUpdate,
					Row:    Row{ID: "A", Fields: Fields{"10", "2", "3"}},
					Deltas: DeltaRow{"1", "", ""},
				},
				RowEvent{Kind: EventUnchanged, Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
				RowEvent{Kind: EventUnchanged, Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
			),
		},
	}

	var table TableData
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			table.SetRowEvents(u.re)
			table.Update(u.rr)
			assert.Equal(t, u.e, table.GetRowEvents())
		})
	}
}

func TestTableDataDelete(t *testing.T) {
	uu := map[string]struct {
		re, e *RowEvents
		kk    map[string]struct{}
	}{
		"ordered": {
			re: NewRowEventsWithEvts(
				RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
				RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
				RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
			),
			kk: map[string]struct{}{"A": {}, "C": {}},
			e: NewRowEventsWithEvts(
				RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
				RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
			),
		},
		"unordered": {
			re: NewRowEventsWithEvts(
				RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
				RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
				RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				RowEvent{Row: Row{ID: "D", Fields: Fields{"10", "2", "3"}}},
			),
			kk: map[string]struct{}{"C": {}, "A": {}},
			e: NewRowEventsWithEvts(
				RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
				RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
			),
		},
	}

	var table TableData
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			table.SetRowEvents(u.re)
			table.Delete(u.kk)
			assert.Equal(t, u.e, table.GetRowEvents())
		})
	}
}
