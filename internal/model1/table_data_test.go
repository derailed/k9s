// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import (
	"log/slog"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/util/sets"
)

func init() {
	slog.SetDefault(slog.New(slog.DiscardHandler))
}

func TestTableDataSortWideColumn(t *testing.T) {
	const (
		nameColumn      = "NAME"
		ownerKindColumn = "OWNER_KIND"
		statefulSetID   = "statefulset"
		daemonSetID     = "daemonset"
		jobID           = "job"
	)

	wideSortColumn := SortColumn{Name: ownerKindColumn, ASC: true}
	expectedIDs := []string{daemonSetID, jobID, statefulSetID}
	newTableData := func() *TableData {
		return NewTableDataWithRows(
			client.NewGVR("test"),
			Header{
				HeaderColumn{Name: nameColumn},
				HeaderColumn{Name: ownerKindColumn, Attrs: Attrs{Wide: true}},
			},
			NewRowEventsWithEvts(
				RowEvent{Row: Row{ID: statefulSetID, Fields: Fields{statefulSetID, "StatefulSet"}}},
				RowEvent{Row: Row{ID: daemonSetID, Fields: Fields{daemonSetID, "DaemonSet"}}},
				RowEvent{Row: Row{ID: jobID, Fields: Fields{jobID, "Job"}}},
			),
		)
	}

	tests := map[string]struct {
		viewSetting        *config.ViewSetting
		initialSortColumn  SortColumn
		manual             bool
		expectedSortColumn SortColumn
		expectedIDs        []string
	}{
		"sorts selected wide column": {
			initialSortColumn:  wideSortColumn,
			expectedSortColumn: wideSortColumn,
			expectedIDs:        expectedIDs,
		},
		"keeps manual sort after wide view is hidden": {
			viewSetting: &config.ViewSetting{
				Columns:    []string{nameColumn, ownerKindColumn},
				SortColumn: "NAME:asc",
			},
			initialSortColumn:  wideSortColumn,
			manual:             true,
			expectedSortColumn: wideSortColumn,
			expectedIDs:        expectedIDs,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			data := newTableData()
			sortColumn := test.initialSortColumn
			if test.viewSetting != nil {
				sortColumn = data.ComputeSortCol(test.viewSetting, sortColumn, test.manual)
			}
			if !assert.Equal(t, test.expectedSortColumn, sortColumn, "unexpected resolved sort column") {
				return
			}

			data.Sort(sortColumn)
			var ids []string
			data.RowsRange(func(_ int, re RowEvent) bool {
				ids = append(ids, re.Row.ID)
				return true
			})
			assert.Equal(t, test.expectedIDs, ids, "unexpected row order after sorting by %q", sortColumn.Name)
		})
	}
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
