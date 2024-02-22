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
		re1, re2 *render.RowEvents
		ageCol   int
		e        bool
	}{
		"same": {
			re1: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			re2: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			ageCol: -1,
		},
		"diff-len": {
			re1: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			re2: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			ageCol: -1,
			e:      true,
		},
		"diff-id": {
			re1: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			re2: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "D", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			ageCol: -1,
			e:      true,
		},
		"diff-order": {
			re1: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			re2: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			ageCol: -1,
			e:      true,
		},
		"diff-withAge": {
			re1: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			re2: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "13"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
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
		ee, e *render.RowEvents
		re    render.RowEvent
	}{
		"add": {
			ee: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			re: render.RowEvent{
				Row: render.Row{ID: "D", Fields: render.Fields{"f1", "f2", "f3"}},
			},
			e: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "D", Fields: render.Fields{"f1", "f2", "f3"}}},
			),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.ee.Upsert(u.re)
			assert.Equal(t, u.e, u.ee)
		})
	}
}

func TestRowEventsCustomize(t *testing.T) {
	uu := map[string]struct {
		re, e *render.RowEvents
		cols  []int
	}{
		"same": {
			re: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			cols: []int{0, 1, 2},
			e: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
		},
		"reverse": {
			re: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			cols: []int{2, 1, 0},
			e: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"3", "2", "1"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"3", "2", "0"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"3", "2", "10"}}},
			),
		},
		"skip": {
			re: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			cols: []int{1, 0},
			e: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"2", "1"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"2", "0"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"2", "10"}}},
			),
		},
		"missing": {
			re: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			cols: []int{1, 0, 4},
			e: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"2", "1", ""}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"2", "0", ""}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"2", "10", ""}}},
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
		re, e *render.RowEvents
		id    string
	}{
		"first": {
			re: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			id: "A",
			e: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
		},
		"middle": {
			re: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			id: "B",
			e: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
		},
		"last": {
			re: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			id: "C",
			e: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
			),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.re.Delete(u.id)
			assert.Equal(t, u.e, u.re)
		})
	}
}

func TestRowEventsSort(t *testing.T) {
	uu := map[string]struct {
		re, e              *render.RowEvents
		col                int
		duration, num, asc bool
		capacity           bool
	}{
		"age_time": {
			re: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", testTime().Add(20 * time.Second).String()}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", testTime().Add(10 * time.Second).String()}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", testTime().String()}}},
			),
			col:      2,
			asc:      true,
			duration: true,
			e: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", testTime().String()}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", testTime().Add(10 * time.Second).String()}}},
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", testTime().Add(20 * time.Second).String()}}},
			),
		},
		"col0": {
			re: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
			col: 0,
			asc: true,
			e: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "B", Fields: render.Fields{"0", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "A", Fields: render.Fields{"1", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "C", Fields: render.Fields{"10", "2", "3"}}},
			),
		},
		"id_preserve": {
			re: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "ns1/B", Fields: render.Fields{"B", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "ns1/A", Fields: render.Fields{"A", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "ns1/C", Fields: render.Fields{"C", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "ns2/B", Fields: render.Fields{"B", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "ns2/A", Fields: render.Fields{"A", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "ns2/C", Fields: render.Fields{"C", "2", "3"}}},
			),
			col: 1,
			asc: true,
			e: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "ns1/A", Fields: render.Fields{"A", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "ns1/B", Fields: render.Fields{"B", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "ns1/C", Fields: render.Fields{"C", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "ns2/A", Fields: render.Fields{"A", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "ns2/B", Fields: render.Fields{"B", "2", "3"}}},
				render.RowEvent{Row: render.Row{ID: "ns2/C", Fields: render.Fields{"C", "2", "3"}}},
			),
		},
		"capacity": {
			re: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "ns1/B", Fields: render.Fields{"B", "2", "3", "1Gi"}}},
				render.RowEvent{Row: render.Row{ID: "ns1/A", Fields: render.Fields{"A", "2", "3", "1.1G"}}},
				render.RowEvent{Row: render.Row{ID: "ns1/C", Fields: render.Fields{"C", "2", "3", "0.5Ti"}}},
				render.RowEvent{Row: render.Row{ID: "ns2/B", Fields: render.Fields{"B", "2", "3", "12e6"}}},
				render.RowEvent{Row: render.Row{ID: "ns2/A", Fields: render.Fields{"A", "2", "3", "1234"}}},
				render.RowEvent{Row: render.Row{ID: "ns2/C", Fields: render.Fields{"C", "2", "3", "0.1Ei"}}},
			),
			col:      3,
			asc:      true,
			capacity: true,
			e: render.NewRowEventsWithEvts(
				render.RowEvent{Row: render.Row{ID: "ns2/A", Fields: render.Fields{"A", "2", "3", "1234"}}},
				render.RowEvent{Row: render.Row{ID: "ns2/B", Fields: render.Fields{"B", "2", "3", "12e6"}}},
				render.RowEvent{Row: render.Row{ID: "ns1/B", Fields: render.Fields{"B", "2", "3", "1Gi"}}},
				render.RowEvent{Row: render.Row{ID: "ns1/A", Fields: render.Fields{"A", "2", "3", "1.1G"}}},
				render.RowEvent{Row: render.Row{ID: "ns1/C", Fields: render.Fields{"C", "2", "3", "0.5Ti"}}},
				render.RowEvent{Row: render.Row{ID: "ns2/C", Fields: render.Fields{"C", "2", "3", "0.1Ei"}}},
			),
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
		r *render.RowEvents
	}{
		"empty": {
			r: render.NewRowEventsWithEvts(),
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

// Helpers...

func makeRowEvents() *render.RowEvents {
	return render.NewRowEventsWithEvts(
		render.RowEvent{Row: render.Row{ID: "ns1/A", Fields: render.Fields{"A", "2", "3"}}},
		render.RowEvent{Row: render.Row{ID: "ns1/B", Fields: render.Fields{"B", "2", "3"}}},
		render.RowEvent{Row: render.Row{ID: "ns1/C", Fields: render.Fields{"C", "2", "3"}}},
		render.RowEvent{Row: render.Row{ID: "ns2/A", Fields: render.Fields{"A", "2", "3"}}},
		render.RowEvent{Row: render.Row{ID: "ns2/B", Fields: render.Fields{"B", "2", "3"}}},
		render.RowEvent{Row: render.Row{ID: "ns2/C", Fields: render.Fields{"C", "2", "3"}}},
	)
}
