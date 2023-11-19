// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"
	"time"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestRowEventCustomize(t *testing.T) {
	uu := map[string]struct {
		re1, e render.RowEvent
		cols   []int
	}{
		"empty": {
			re1: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
			},
			e: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{}},
			},
		},
		"full": {
			re1: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
			},
			e: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
			},
			cols: []int{0, 1, 2},
		},
		"deltas": {
			re1: render.RowEvent{
				Kind:   render.EventAdd,
				Row:    render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
				Deltas: render.DeltaRow{"a", "b", "c"},
			},
			e: render.RowEvent{
				Kind:   render.EventAdd,
				Row:    render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
				Deltas: render.DeltaRow{"a", "b", "c"},
			},
			cols: []int{0, 1, 2},
		},
		"deltas-skip": {
			re1: render.RowEvent{
				Kind:   render.EventAdd,
				Row:    render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
				Deltas: render.DeltaRow{"a", "b", "c"},
			},
			e: render.RowEvent{
				Kind:   render.EventAdd,
				Row:    render.Row{ID: "A", Fields: render.Fields{"3", "1"}},
				Deltas: render.DeltaRow{"c", "a"},
			},
			cols: []int{2, 0},
		},
		"reverse": {
			re1: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
			},
			e: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{"3", "2", "1"}},
			},
			cols: []int{2, 1, 0},
		},
		"skip": {
			re1: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
			},
			e: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{"3", "1"}},
			},
			cols: []int{2, 0},
		},
		"miss": {
			re1: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
			},
			e: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{"3", "", "1"}},
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
		re1, re2 render.RowEvent
		e        bool
	}{
		"same": {
			re1: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
			},
			re2: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
			},
		},
		"diff-kind": {
			re1: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
			},
			re2: render.RowEvent{
				Kind: render.EventDelete,
				Row:  render.Row{ID: "B", Fields: render.Fields{"1", "2", "3"}},
			},
			e: true,
		},
		"diff-delta": {
			re1: render.RowEvent{
				Kind:   render.EventAdd,
				Row:    render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
				Deltas: render.DeltaRow{"1", "2", "3"},
			},
			re2: render.RowEvent{
				Kind:   render.EventAdd,
				Row:    render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
				Deltas: render.DeltaRow{"10", "2", "3"},
			},
			e: true,
		},
		"diff-id": {
			re1: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
			},
			re2: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "B", Fields: render.Fields{"1", "2", "3"}},
			},
			e: true,
		},
		"diff-field": {
			re1: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}},
			},
			re2: render.RowEvent{
				Kind: render.EventAdd,
				Row:  render.Row{ID: "A", Fields: render.Fields{"10", "2", "3"}},
			},
			e: true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.re1.Diff(u.re2, -1))
		})
	}
}

func TestRowEventsDiff(t *testing.T) {
	uu := map[string]struct {
		re1, re2 render.RowEvents
		ageCol   int
		e        bool
	}{
		"same": {
			re1: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			re2: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			ageCol: -1,
		},
		"diff-len": {
			re1: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			re2: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			ageCol: -1,
			e:      true,
		},
		"diff-id": {
			re1: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			re2: render.RowEvents{
				{Row: render.Row{ID: "D", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			ageCol: -1,
			e:      true,
		},
		"diff-order": {
			re1: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			re2: render.RowEvents{
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			ageCol: -1,
			e:      true,
		},
		"diff-withAge": {
			re1: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			re2: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "13"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
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
		ee, e render.RowEvents
		re    render.RowEvent
	}{
		"add": {
			ee: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			re: render.RowEvent{
				Row: render.Row{ID: "D", Fields: render.Fields{"f1", "f2", "f3"}},
			},
			e: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				{Row: render.Row{ID: "D", Fields: render.Fields{"f1", "f2", "f3"}}},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.ee.Upsert(u.re))
		})
	}
}

func TestRowEventsCustomize(t *testing.T) {
	uu := map[string]struct {
		re, e render.RowEvents
		cols  []int
	}{
		"same": {
			re: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			cols: []int{0, 1, 2},
			e: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
		},
		"reverse": {
			re: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			cols: []int{2, 1, 0},
			e: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"3", "2", "1"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"3", "2", "0"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"3", "2", "10"}}},
			},
		},
		"skip": {
			re: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			cols: []int{1, 0},
			e: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"2", "1"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"2", "0"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"2", "10"}}},
			},
		},
		"missing": {
			re: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			},
			cols: []int{1, 0, 4},
			e: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"2", "1", ""}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"2", "0", ""}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"2", "10", ""}}},
			},
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

func TestRowEventsSort(t *testing.T) {
	uu := map[string]struct {
		re                 render.RowEvents
		col                int
		duration, num, asc bool
		capacity           bool
		e                  render.RowEvents
	}{
		"age_time": {
			re: render.RowEvents{
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", testTime().Add(20 * time.Second).String()}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", testTime().Add(10 * time.Second).String()}}},
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", testTime().String()}}},
			},
			col:      2,
			asc:      true,
			duration: true,
			e: render.RowEvents{
				{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", testTime().String()}}},
				{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", testTime().Add(10 * time.Second).String()}}},
				{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", testTime().Add(20 * time.Second).String()}}},
			},
		},
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
		"capacity": {
			re: render.RowEvents{
				{Row: render.Row{ID: "ns1/B", Fields: render.Fields{"B", "2", "3", "1Gi"}}},
				{Row: render.Row{ID: "ns1/A", Fields: render.Fields{"A", "2", "3", "1.1G"}}},
				{Row: render.Row{ID: "ns1/C", Fields: render.Fields{"C", "2", "3", "0.5Ti"}}},
				{Row: render.Row{ID: "ns2/B", Fields: render.Fields{"B", "2", "3", "12e6"}}},
				{Row: render.Row{ID: "ns2/A", Fields: render.Fields{"A", "2", "3", "1234"}}},
				{Row: render.Row{ID: "ns2/C", Fields: render.Fields{"C", "2", "3", "0.1Ei"}}},
			},
			col:      3,
			asc:      true,
			capacity: true,
			e: render.RowEvents{
				{Row: render.Row{ID: "ns2/A", Fields: render.Fields{"A", "2", "3", "1234"}}},
				{Row: render.Row{ID: "ns2/B", Fields: render.Fields{"B", "2", "3", "12e6"}}},
				{Row: render.Row{ID: "ns1/B", Fields: render.Fields{"B", "2", "3", "1Gi"}}},
				{Row: render.Row{ID: "ns1/A", Fields: render.Fields{"A", "2", "3", "1.1G"}}},
				{Row: render.Row{ID: "ns1/C", Fields: render.Fields{"C", "2", "3", "0.5Ti"}}},
				{Row: render.Row{ID: "ns2/C", Fields: render.Fields{"C", "2", "3", "0.1Ei"}}},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.re.Sort("", u.col, u.duration, u.num, u.capacity, u.asc)
			assert.Equal(t, u.e, u.re)
		})
	}
}

func TestRowEventsClone(t *testing.T) {
	uu := map[string]struct {
		r render.RowEvents
	}{
		"empty": {
			r: render.RowEvents{},
		},
		"full": {
			r: makeRowEvents(),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := u.r.Clone()
			assert.Equal(t, len(u.r), len(c))
			if len(u.r) > 0 {
				u.r[0].Row.Fields[0] = "blee"
				assert.Equal(t, "A", c[0].Row.Fields[0])
			}
		})
	}
}

// Helpers...

func makeRowEvents() render.RowEvents {
	return render.RowEvents{
		{Row: render.Row{ID: "ns1/A", Fields: render.Fields{"A", "2", "3"}}},
		{Row: render.Row{ID: "ns1/B", Fields: render.Fields{"B", "2", "3"}}},
		{Row: render.Row{ID: "ns1/C", Fields: render.Fields{"C", "2", "3"}}},
		{Row: render.Row{ID: "ns2/A", Fields: render.Fields{"A", "2", "3"}}},
		{Row: render.Row{ID: "ns2/B", Fields: render.Fields{"B", "2", "3"}}},
		{Row: render.Row{ID: "ns2/C", Fields: render.Fields{"C", "2", "3"}}},
	}
}
