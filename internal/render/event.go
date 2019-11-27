package render

import (
	"sort"

	"github.com/gdamore/tcell"
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

// Delete removes an element by id.
func (re RowEvents) Delete(id string) RowEvents {
	idx, ok := re.FindIndex(id)
	if !ok {
		return re
	}

	if idx == 0 {
		return re[1:]
	}
	if idx == len(re)-1 {
		return re[:len(re)-1]
	}

	return append(re[:idx], re[idx+1:]...)
}

// FindIndex locates a row index by id. Returns false is not found.
func (re RowEvents) FindIndex(id string) (int, bool) {
	for i, e := range re {
		if e.Row.ID == id {
			return i, true
		}
	}

	return 0, false
}

// Sort rows based on column index and order.
func (re RowEvents) Sort(ns string, col int, asc bool) {
	t := RowEventSorter{NS: ns, Events: re, Index: col, Asc: asc}
	sort.Sort(t)
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

	var col int
	if r.NS == "" {
		col++
	}
	if col >= len(f1) || col >= len(f2) {
		return false
	}
	n1, n2 := f1[col], f2[col]

	return Less(r.Asc, f1[r.Index]+n1, f2[r.Index]+n2)
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
type ColorerFunc func(ns string, evt ResEvent, r Row) tcell.Color

// DefaultColorer set the default table row colors.
func DefaultColorer(ns string, evt ResEvent, r Row) tcell.Color {
	switch evt {
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
