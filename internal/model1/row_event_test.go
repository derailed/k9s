// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1_test

import (
	"testing"
	"time"

	"github.com/derailed/k9s/internal/model1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRowEventCustomize(t *testing.T) {
	uu := map[string]struct {
		re1, e model1.RowEvent
		cols   []int
	}{
		"empty": {
			re1: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
			},
			e: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{}},
			},
		},
		"full": {
			re1: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
			},
			e: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
			},
			cols: []int{0, 1, 2},
		},
		"deltas": {
			re1: model1.RowEvent{
				Kind:   model1.EventAdd,
				Row:    model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
				Deltas: model1.DeltaRow{"a", "b", "c"},
			},
			e: model1.RowEvent{
				Kind:   model1.EventAdd,
				Row:    model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
				Deltas: model1.DeltaRow{"a", "b", "c"},
			},
			cols: []int{0, 1, 2},
		},
		"deltas-skip": {
			re1: model1.RowEvent{
				Kind:   model1.EventAdd,
				Row:    model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
				Deltas: model1.DeltaRow{"a", "b", "c"},
			},
			e: model1.RowEvent{
				Kind:   model1.EventAdd,
				Row:    model1.Row{ID: "A", Fields: model1.Fields{"3", "1"}},
				Deltas: model1.DeltaRow{"c", "a"},
			},
			cols: []int{2, 0},
		},
		"reverse": {
			re1: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
			},
			e: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{"3", "2", "1"}},
			},
			cols: []int{2, 1, 0},
		},
		"skip": {
			re1: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
			},
			e: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{"3", "1"}},
			},
			cols: []int{2, 0},
		},
		"miss": {
			re1: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
			},
			e: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{"3", "", "1"}},
			},
			cols: []int{2, 10, 0},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.re1.Customize(u.cols))
		})
	}
}

func TestRowEventDiff(t *testing.T) {
	uu := map[string]struct {
		re1, re2 model1.RowEvent
		e        bool
	}{
		"same": {
			re1: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
			},
			re2: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
			},
		},
		"diff-kind": {
			re1: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
			},
			re2: model1.RowEvent{
				Kind: model1.EventDelete,
				Row:  model1.Row{ID: "B", Fields: model1.Fields{"1", "2", "3"}},
			},
			e: true,
		},
		"diff-delta": {
			re1: model1.RowEvent{
				Kind:   model1.EventAdd,
				Row:    model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
				Deltas: model1.DeltaRow{"1", "2", "3"},
			},
			re2: model1.RowEvent{
				Kind:   model1.EventAdd,
				Row:    model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
				Deltas: model1.DeltaRow{"10", "2", "3"},
			},
			e: true,
		},
		"diff-id": {
			re1: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
			},
			re2: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "B", Fields: model1.Fields{"1", "2", "3"}},
			},
			e: true,
		},
		"diff-field": {
			re1: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}},
			},
			re2: model1.RowEvent{
				Kind: model1.EventAdd,
				Row:  model1.Row{ID: "A", Fields: model1.Fields{"10", "2", "3"}},
			},
			e: true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.re1.Diff(&u.re2, -1))
		})
	}
}

func TestRowEventsDiff(t *testing.T) {
	uu := map[string]struct {
		re1, re2 *model1.RowEvents
		ageCol   int
		e        bool
	}{
		"same": {
			re1: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			re2: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			ageCol: -1,
		},
		"diff-len": {
			re1: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			re2: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			ageCol: -1,
			e:      true,
		},
		"diff-id": {
			re1: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			re2: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "D", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			ageCol: -1,
			e:      true,
		},
		"diff-order": {
			re1: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			re2: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			ageCol: -1,
			e:      true,
		},
		"diff-withAge": {
			re1: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			re2: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "13"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			ageCol: 1,
			e:      true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.re1.Diff(u.re2, u.ageCol))
		})
	}
}

func TestRowEventsUpsert(t *testing.T) {
	uu := map[string]struct {
		ee, e *model1.RowEvents
		re    model1.RowEvent
	}{
		"add": {
			ee: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			re: model1.RowEvent{
				Row: model1.Row{ID: "D", Fields: model1.Fields{"f1", "f2", "f3"}},
			},
			e: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "D", Fields: model1.Fields{"f1", "f2", "f3"}}},
			),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.ee.Upsert(&u.re)
			assert.Equal(t, u.e, u.ee)
		})
	}
}

func TestRowEventsCustomize(t *testing.T) {
	uu := map[string]struct {
		re, e *model1.RowEvents
		cols  []int
	}{
		"same": {
			re: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			cols: []int{0, 1, 2},
			e: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
		},
		"reverse": {
			re: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			cols: []int{2, 1, 0},
			e: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"3", "2", "1"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"3", "2", "0"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"3", "2", "10"}}},
			),
		},
		"skip": {
			re: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			cols: []int{1, 0},
			e: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"2", "1"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"2", "0"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"2", "10"}}},
			),
		},
		"missing": {
			re: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			cols: []int{1, 0, 4},
			e: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"2", "1", ""}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"2", "0", ""}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"2", "10", ""}}},
			),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.re.Customize(u.cols))
		})
	}
}

func TestRowEventsDelete(t *testing.T) {
	uu := map[string]struct {
		re, e *model1.RowEvents
		id    string
	}{
		"first": {
			re: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			id: "A",
			e: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
		},
		"middle": {
			re: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			id: "B",
			e: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
		},
		"last": {
			re: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			id: "C",
			e: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
			),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			require.NoError(t, u.re.Delete(u.id))
			assert.Equal(t, u.e, u.re)
		})
	}
}

func TestRowEventsSort(t *testing.T) {
	uu := map[string]struct {
		re, e              *model1.RowEvents
		col                int
		duration, num, asc bool
		capacity           bool
	}{
		"age_time": {
			re: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", testTime().Add(20 * time.Second).String()}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", testTime().Add(10 * time.Second).String()}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", testTime().String()}}},
			),
			col:      2,
			asc:      true,
			duration: true,
			e: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", testTime().String()}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", testTime().Add(10 * time.Second).String()}}},
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", testTime().Add(20 * time.Second).String()}}},
			),
		},
		"col0": {
			re: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
			col: 0,
			asc: true,
			e: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "B", Fields: model1.Fields{"0", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "A", Fields: model1.Fields{"1", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "C", Fields: model1.Fields{"10", "2", "3"}}},
			),
		},
		"id_preserve": {
			re: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "ns1/B", Fields: model1.Fields{"B", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns1/A", Fields: model1.Fields{"A", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns1/C", Fields: model1.Fields{"C", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns2/B", Fields: model1.Fields{"B", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns2/A", Fields: model1.Fields{"A", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns2/C", Fields: model1.Fields{"C", "2", "3"}}},
			),
			col: 1,
			asc: true,
			e: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "ns1/A", Fields: model1.Fields{"A", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns1/B", Fields: model1.Fields{"B", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns1/C", Fields: model1.Fields{"C", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns2/A", Fields: model1.Fields{"A", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns2/B", Fields: model1.Fields{"B", "2", "3"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns2/C", Fields: model1.Fields{"C", "2", "3"}}},
			),
		},
		"capacity": {
			re: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "ns1/B", Fields: model1.Fields{"B", "2", "3", "1Gi"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns1/A", Fields: model1.Fields{"A", "2", "3", "1.1G"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns1/C", Fields: model1.Fields{"C", "2", "3", "0.5Ti"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns2/B", Fields: model1.Fields{"B", "2", "3", "12e6"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns2/A", Fields: model1.Fields{"A", "2", "3", "1234"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns2/C", Fields: model1.Fields{"C", "2", "3", "0.1Ei"}}},
			),
			col:      3,
			asc:      true,
			capacity: true,
			e: model1.NewRowEventsWithEvts(
				model1.RowEvent{Row: model1.Row{ID: "ns2/A", Fields: model1.Fields{"A", "2", "3", "1234"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns2/B", Fields: model1.Fields{"B", "2", "3", "12e6"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns1/B", Fields: model1.Fields{"B", "2", "3", "1Gi"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns1/A", Fields: model1.Fields{"A", "2", "3", "1.1G"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns1/C", Fields: model1.Fields{"C", "2", "3", "0.5Ti"}}},
				model1.RowEvent{Row: model1.Row{ID: "ns2/C", Fields: model1.Fields{"C", "2", "3", "0.1Ei"}}},
			),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.re.Sort("", u.col, "", u.duration, u.num, u.capacity, u.asc)
			assert.Equal(t, u.e, u.re)
		})
	}
}

func TestRowEventsClone(t *testing.T) {
	uu := map[string]struct {
		r *model1.RowEvents
	}{
		"empty": {
			r: model1.NewRowEventsWithEvts(),
		},
		"full": {
			r: makeRowEvents(),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := u.r.Clone()
			assert.Equal(t, u.r.Len(), c.Len())
			if !u.r.Empty() {
				r, ok := u.r.At(0)
				assert.True(t, ok)
				r.Row.Fields[0] = "blee"
				cr, ok := c.At(0)
				assert.True(t, ok)
				assert.Equal(t, "A", cr.Row.Fields[0])
			}
		})
	}
}

func TestRowEventsSortTimestampStability(t *testing.T) {
	now := time.Now()
	// Six pods all showing "12m" as humanized age, but with different actual creation times.
	mkRow := func(id string, createdAgo time.Duration) model1.Row {
		return model1.Row{
			ID:         id,
			Fields:     model1.Fields{id, "Running", "12m"},
			Timestamps: map[string]time.Time{"AGE": now.Add(-createdAgo)},
		}
	}

	events := model1.NewRowEventsWithEvts(
		model1.NewRowEvent(model1.EventAdd, mkRow("ns/f", 12*time.Minute+50*time.Second)),
		model1.NewRowEvent(model1.EventAdd, mkRow("ns/a", 12*time.Minute+10*time.Second)),
		model1.NewRowEvent(model1.EventAdd, mkRow("ns/d", 12*time.Minute+40*time.Second)),
		model1.NewRowEvent(model1.EventAdd, mkRow("ns/b", 12*time.Minute+20*time.Second)),
		model1.NewRowEvent(model1.EventAdd, mkRow("ns/e", 12*time.Minute+45*time.Second)),
		model1.NewRowEvent(model1.EventAdd, mkRow("ns/c", 12*time.Minute+30*time.Second)),
	)

	// Sort ascending by column 2 (AGE) — a duration/time column.
	events.Sort("", 2, "AGE", true, false, false, true)

	// Expected ascending order: newest (smallest age) first.
	// ns/a (12m10s) < ns/b (12m20s) < ns/c (12m30s) < ns/d (12m40s) < ns/e (12m45s) < ns/f (12m50s)
	expectedIDs := []string{"ns/a", "ns/b", "ns/c", "ns/d", "ns/e", "ns/f"}
	gotIDs := make([]string, events.Len())
	events.Range(func(i int, re model1.RowEvent) bool {
		gotIDs[i] = re.Row.ID
		return true
	})
	assert.Equal(t, expectedIDs, gotIDs, "ascending timestamp sort")

	// Sort again (simulates refresh) — order must not change.
	events.Sort("", 2, "AGE", true, false, false, true)
	gotIDs2 := make([]string, events.Len())
	events.Range(func(i int, re model1.RowEvent) bool {
		gotIDs2[i] = re.Row.ID
		return true
	})
	assert.Equal(t, expectedIDs, gotIDs2, "repeated ascending sort must be stable")

	// Sort descending — oldest (largest age) first.
	events.Sort("", 2, "AGE", true, false, false, false)
	expectedDesc := []string{"ns/f", "ns/e", "ns/d", "ns/c", "ns/b", "ns/a"}
	gotDesc := make([]string, events.Len())
	events.Range(func(i int, re model1.RowEvent) bool {
		gotDesc[i] = re.Row.ID
		return true
	})
	assert.Equal(t, expectedDesc, gotDesc, "descending timestamp sort")
}

func TestRowEventsSortTimestampFallback(t *testing.T) {
	// Rows WITHOUT stashed timestamps — must fall back to string-based sort.
	events := model1.NewRowEventsWithEvts(
		model1.NewRowEvent(model1.EventAdd, model1.Row{ID: "ns/c", Fields: model1.Fields{"c", "Running", "12m"}}),
		model1.NewRowEvent(model1.EventAdd, model1.Row{ID: "ns/a", Fields: model1.Fields{"a", "Running", "12m"}}),
		model1.NewRowEvent(model1.EventAdd, model1.Row{ID: "ns/b", Fields: model1.Fields{"b", "Running", "12m"}}),
	)

	events.Sort("", 2, "AGE", true, false, false, true)

	// All have the same duration string "12m", so fallback sorts by ID.
	got := make([]string, events.Len())
	events.Range(func(i int, re model1.RowEvent) bool {
		got[i] = re.Row.ID
		return true
	})
	assert.Equal(t, []string{"ns/a", "ns/b", "ns/c"}, got, "fallback sort by ID when ages equal")
}

// Helpers...

func makeRowEvents() *model1.RowEvents {
	return model1.NewRowEventsWithEvts(
		model1.RowEvent{Row: model1.Row{ID: "ns1/A", Fields: model1.Fields{"A", "2", "3"}}},
		model1.RowEvent{Row: model1.Row{ID: "ns1/B", Fields: model1.Fields{"B", "2", "3"}}},
		model1.RowEvent{Row: model1.Row{ID: "ns1/C", Fields: model1.Fields{"C", "2", "3"}}},
		model1.RowEvent{Row: model1.Row{ID: "ns2/A", Fields: model1.Fields{"A", "2", "3"}}},
		model1.RowEvent{Row: model1.Row{ID: "ns2/B", Fields: model1.Fields{"B", "2", "3"}}},
		model1.RowEvent{Row: model1.Row{ID: "ns2/C", Fields: model1.Fields{"C", "2", "3"}}},
	)
}
