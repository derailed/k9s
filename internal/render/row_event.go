package render

import (
	"fmt"
	"reflect"
	"sort"

	"github.com/rs/zerolog/log"
)

const (
	// EventUnchanged notifies listener resource has not changed.
	EventUnchanged ResEvent = 1 << iota

	// EventAdd notifies listener of a resource was added.
	EventAdd

	// EventUpdate notifies listener of a resource updated.
	EventUpdate

	// EventDelete  notifies listener of a resource was deleted.
	EventDelete

	// EventClear the stack was reset.
	EventClear
)

// ResEvent represents a resource event.
type ResEvent int

// RowEvent tracks resource instance events.
type RowEvent struct {
	Kind   ResEvent
	Row    Row
	Deltas DeltaRow
}

// NewRowEvent returns a new row event.
func NewRowEvent(kind ResEvent, row Row) RowEvent {
	return RowEvent{
		Kind: kind,
		Row:  row,
	}
}

// NewDeltaRowEvent returns a new row event with deltas.
func NewDeltaRowEvent(row Row, delta DeltaRow) RowEvent {
	return RowEvent{
		Kind:   EventUpdate,
		Row:    row,
		Deltas: delta,
	}
}

// Clone returns a rowevent deep copy.
func (r RowEvent) Clone() RowEvent {
	return RowEvent{
		Kind:   r.Kind,
		Row:    r.Row.Clone(),
		Deltas: r.Deltas.Clone(),
	}
}

// Changed returns true if the row changed.
func (r RowEvent) Changed(re RowEvent) bool {
	if r.Kind != re.Kind {
		log.Debug().Msgf("KIND Changed")
		return true
	}
	if !reflect.DeepEqual(r.Deltas, re.Deltas) {
		log.Debug().Msgf("DELTAS CHANGED")
		return true
	}

	return !reflect.DeepEqual(r.Row.Fields[:len(r.Row.Fields)-1], re.Row.Fields[:len(re.Row.Fields)-1])
}

// ----------------------------------------------------------------------------

// RowEvents a collection of row events.
type RowEvents []RowEvent

// Changed returns true if the header changed.
func (rr RowEvents) Changed(r RowEvents) bool {
	if len(rr) != len(r) {
		return true
	}

	for i := range rr {
		if rr[i].Changed(r[i]) {
			return true
		}
	}

	return false
}

// Clone returns a rowevents deep copy.
func (rr RowEvents) Clone() RowEvents {
	res := make(RowEvents, len(rr))
	for i, r := range rr {
		res[i] = r.Clone()
	}

	return res
}

// Upsert add or update a row if it exists.
func (rr RowEvents) Upsert(e RowEvent) RowEvents {
	if idx, ok := rr.FindIndex(e.Row.ID); ok {
		rr[idx] = e
	} else {
		rr = append(rr, e)
	}
	return rr
}

// Delete removes an element by id.
func (rr RowEvents) Delete(id string) RowEvents {
	idx, ok := rr.FindIndex(id)
	if !ok {
		return rr
	}

	if idx == 0 {
		return rr[1:]
	}
	if idx == len(rr)-1 {
		return rr[:len(rr)-1]
	}

	return append(rr[:idx], rr[idx+1:]...)
}

// Clear delete all row events
func (rr RowEvents) Clear() RowEvents {
	return RowEvents{}
}

// FindIndex locates a row index by id. Returns false is not found.
func (rr RowEvents) FindIndex(id string) (int, bool) {
	for i, e := range rr {
		if e.Row.ID == id {
			return i, true
		}
	}

	return 0, false
}

// Sort rows based on column index and order.
func (rr RowEvents) Sort(ns string, col int, asc bool) {
	t := RowEventSorter{NS: ns, Events: rr, Index: col, Asc: asc}
	sort.Sort(t)

	gg, kk := map[string][]string{}, make(StringSet, 0, len(rr))
	for _, e := range rr {
		g := e.Row.Fields[col]
		kk = kk.Add(g)
		if ss, ok := gg[g]; ok {
			gg[g] = append(ss, e.Row.ID)
		} else {
			gg[g] = []string{e.Row.ID}
		}
	}

	ids := make([]string, 0, len(rr))
	for _, k := range kk {
		sort.StringSlice(gg[k]).Sort()
		ids = append(ids, gg[k]...)
	}
	s := IdSorter{Ids: ids, Events: rr}
	sort.Sort(s)
}

// ----------------------------------------------------------------------------

// RowEventSorter sorts row events by a given colon.
type RowEventSorter struct {
	Events RowEvents
	Index  int
	NS     string
	Asc    bool
}

func (r RowEventSorter) Len() int {
	return len(r.Events)
}

func (r RowEventSorter) Swap(i, j int) {
	r.Events[i], r.Events[j] = r.Events[j], r.Events[i]
}

func (r RowEventSorter) Less(i, j int) bool {
	f1, f2 := r.Events[i].Row.Fields, r.Events[j].Row.Fields
	return Less(r.Asc, f1[r.Index], f2[r.Index])
}

// ----------------------------------------------------------------------------

// IdSorter sorts row events by a given id.
type IdSorter struct {
	Ids    []string
	Events RowEvents
}

func (s IdSorter) Len() int {
	return len(s.Events)
}

func (s IdSorter) Swap(i, j int) {
	s.Events[i], s.Events[j] = s.Events[j], s.Events[i]
}

func (s IdSorter) Less(i, j int) bool {
	id1, id2 := s.Events[i].Row.ID, s.Events[j].Row.ID
	i1, i2 := findIndex(s.Ids, id1), findIndex(s.Ids, id2)
	return i1 < i2
}

func findIndex(ss []string, s string) int {
	for i := range ss {
		if ss[i] == s {
			return i
		}
	}
	log.Error().Err(fmt.Errorf("Doh! index not found for %s", s))
	return -1
}

// ----------------------------------------------------------------------------

// StringSet represents a collection of unique strings.
type StringSet []string

// Add adds a new item in the set.
func (ss StringSet) Add(item string) StringSet {
	if ss.In(item) {
		return ss
	}
	return append(ss, item)
}

// In checks if a string is in the set.
func (ss StringSet) In(item string) bool {
	return ss.indexOf(item) >= 0
}

func (ss StringSet) indexOf(item string) int {
	for i, s := range ss {
		if s == item {
			return i
		}
	}
	return -1
}
