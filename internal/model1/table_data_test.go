// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import (
	"log/slog"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"
)

func init() {
	slog.SetDefault(slog.New(slog.DiscardHandler))
}

func TestTableDataComputeSortCol(t *testing.T) {
	uu := map[string]struct {
		t1           *TableData
		vs           config.ViewSetting
		sc           SortColumn
		wide, manual bool
		e            SortColumn
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
			vs: config.ViewSetting{Columns: []string{"A", "B", "C"}, SortColumn: "A:asc"},
			e:  SortColumn{Name: "A", ASC: true},
		},
		"wide-col": {
			t1: NewTableDataWithRows(
				client.NewGVR("test"),
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "B", Attrs: Attrs{Wide: true}},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				),
			),
			vs: config.ViewSetting{Columns: []string{"A", "B", "C"}, SortColumn: "B:desc"},
			e:  SortColumn{Name: "B"},
		},

		"wide": {
			t1: NewTableDataWithRows(
				client.NewGVR("test"),
				Header{
					HeaderColumn{Name: "A"},
					HeaderColumn{Name: "B", Attrs: Attrs{Wide: true}},
					HeaderColumn{Name: "C"},
				},
				NewRowEventsWithEvts(
					RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
					RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
					RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
				),
			),
			wide: true,
			vs:   config.ViewSetting{Columns: []string{"A", "C"}, SortColumn: ""},
			e:    SortColumn{Name: ""},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			sc := u.t1.ComputeSortCol(&u.vs, u.sc, u.manual)
			assert.Equal(t, u.e, sc)
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
		kk    sets.Set[string]
	}{
		"ordered": {
			re: NewRowEventsWithEvts(
				RowEvent{Row: Row{ID: "A", Fields: Fields{"1", "2", "3"}}},
				RowEvent{Row: Row{ID: "B", Fields: Fields{"0", "2", "3"}}},
				RowEvent{Row: Row{ID: "C", Fields: Fields{"10", "2", "3"}}},
			),
			kk: sets.New[string]("A", "C"),
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
			kk: sets.New[string]("C", "A"),
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

func TestTableDataTimestampSortPipeline(t *testing.T) {
	now := time.Now()
	gvr := client.NewGVR("v1/pods")
	header := Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "STATUS"},
		HeaderColumn{Name: "AGE", Attrs: Attrs{Time: true}},
	}

	// Simulate initial Hydrate: rows with timestamps (like renderers produce).
	rows := Rows{
		{
			ID:         "ns/pod-c",
			Fields:     Fields{"ns", "pod-c", "Running", "12m"},
			Timestamps: map[string]time.Time{"AGE": now.Add(-12*time.Minute - 30*time.Second)},
		},
		{
			ID:         "ns/pod-a",
			Fields:     Fields{"ns", "pod-a", "Running", "12m"},
			Timestamps: map[string]time.Time{"AGE": now.Add(-12*time.Minute - 10*time.Second)},
		},
		{
			ID:         "ns/pod-b",
			Fields:     Fields{"ns", "pod-b", "Running", "12m"},
			Timestamps: map[string]time.Time{"AGE": now.Add(-12*time.Minute - 20*time.Second)},
		},
	}

	// Step 1: TableData.Update (initial load).
	td := NewTableData(gvr)
	td.SetHeader("ns", header)
	td.Update(rows)

	// Step 2: Peek (clone).
	cloned := td.Clone()

	// Step 3: Sort by AGE ascending.
	cloned.Sort(SortColumn{Name: "AGE", ASC: true})

	// Verify: newest (smallest age) first.
	gotIDs := make([]string, 0, 3)
	cloned.RowsRange(func(_ int, re RowEvent) bool {
		gotIDs = append(gotIDs, re.Row.ID)
		return true
	})
	assert.Equal(t, []string{"ns/pod-a", "ns/pod-b", "ns/pod-c"}, gotIDs,
		"first sort: newest first ascending")

	// Step 4: Simulate refresh — same humanized age, exact same rows.
	td.Update(rows)
	cloned2 := td.Clone()
	cloned2.Sort(SortColumn{Name: "AGE", ASC: true})

	gotIDs2 := make([]string, 0, 3)
	cloned2.RowsRange(func(_ int, re RowEvent) bool {
		gotIDs2 = append(gotIDs2, re.Row.ID)
		return true
	})
	assert.Equal(t, gotIDs, gotIDs2, "second sort after refresh: order must be identical")

	// Step 5: Simulate age boundary crossing — pod-a goes to "13m".
	updatedRows := Rows{
		{
			ID:         "ns/pod-c",
			Fields:     Fields{"ns", "pod-c", "Running", "12m"},
			Timestamps: map[string]time.Time{"AGE": now.Add(-12*time.Minute - 30*time.Second)},
		},
		{
			ID:         "ns/pod-a",
			Fields:     Fields{"ns", "pod-a", "Running", "13m"},
			Timestamps: map[string]time.Time{"AGE": now.Add(-12*time.Minute - 10*time.Second)},
		},
		{
			ID:         "ns/pod-b",
			Fields:     Fields{"ns", "pod-b", "Running", "12m"},
			Timestamps: map[string]time.Time{"AGE": now.Add(-12*time.Minute - 20*time.Second)},
		},
	}
	td.Update(updatedRows)
	cloned3 := td.Clone()
	cloned3.Sort(SortColumn{Name: "AGE", ASC: true})

	gotIDs3 := make([]string, 0, 3)
	cloned3.RowsRange(func(_ int, re RowEvent) bool {
		gotIDs3 = append(gotIDs3, re.Row.ID)
		return true
	})
	// Despite pod-a's display changing to "13m", its ACTUAL timestamp is still newest.
	assert.Equal(t, []string{"ns/pod-a", "ns/pod-b", "ns/pod-c"}, gotIDs3,
		"after age boundary crossing: timestamp sort still correct")
}
