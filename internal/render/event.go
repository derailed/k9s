package render

import (
	"fmt"
	"sort"

	"github.com/gdamore/tcell"
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

// RowEvents a collection of row events.
type RowEvents []RowEvent

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
	for _, e := range rr {
		rr = rr.Delete(e.Row.ID)
	}
	return rr
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

var (
	// ModColor row modified color.
	ModColor tcell.Color
	// AddColor row added color.
	AddColor tcell.Color
	// ErrColor row err color.
	ErrColor tcell.Color
	// StdColor row default color.
	StdColor tcell.Color
	// HighlightColor row highlight color.
	HighlightColor tcell.Color
	// KillColor row deleted color.
	KillColor tcell.Color
	// CompletedColor row completed color.
	CompletedColor tcell.Color
)

// ColorerFunc represents a resource row colorer.
type ColorerFunc func(ns string, evt RowEvent) tcell.Color

// DefaultColorer set the default table row colors.
func DefaultColorer(ns string, evt RowEvent) tcell.Color {
	switch evt.Kind {
	case EventAdd:
		return AddColor
	case EventUpdate:
		return ModColor
	case EventDelete:
		return KillColor
	default:
		return StdColor
	}
}

type StringSet []string

func (ss StringSet) Add(item string) StringSet {
	if ss.In(item) {
		return ss
	}
	return append(ss, item)
}

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
