package render

import (
	"sort"
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

// NewRowEventWithDeltas returns a new row event with deltas.
func NewRowEventWithDeltas(row Row, delta DeltaRow) RowEvent {
	return RowEvent{
		Kind:   EventUpdate,
		Row:    row,
		Deltas: delta,
	}
}

// Clone returns a row event deep copy.
func (r RowEvent) Clone() RowEvent {
	return RowEvent{
		Kind:   r.Kind,
		Row:    r.Row.Clone(),
		Deltas: r.Deltas.Clone(),
	}
}

// Customize returns a new subset based on the given column indices.
func (r RowEvent) Customize(cols []int) RowEvent {
	delta := r.Deltas
	if !r.Deltas.IsBlank() {
		delta = make(DeltaRow, len(cols))
		r.Deltas.Customize(cols, delta)
	}

	return RowEvent{
		Kind:   r.Kind,
		Deltas: delta,
		Row:    r.Row.Customize(cols),
	}
}

// ExtractHeaderLabels extract collection of fields into header.
func (r RowEvent) ExtractHeaderLabels(labelCol int) []string {
	hh, _ := sortLabels(labelize(r.Row.Fields[labelCol]))
	return hh
}

// Labelize returns a new row event based on labels.
func (r RowEvent) Labelize(cols []int, labelCol int, labels []string) RowEvent {
	return RowEvent{
		Kind:   r.Kind,
		Deltas: r.Deltas.Labelize(cols, labelCol),
		Row:    r.Row.Labelize(cols, labelCol, labels),
	}
}

// Diff returns true if the row changed.
func (r RowEvent) Diff(re RowEvent, ageCol int) bool {
	if r.Kind != re.Kind {
		return true
	}
	if r.Deltas.Diff(re.Deltas, ageCol) {
		return true
	}
	return r.Row.Diff(re.Row, ageCol)
}

// ----------------------------------------------------------------------------

// RowEvents a collection of row events.
type RowEvents []RowEvent

// ExtractHeaderLabels extract header labels.
func (r RowEvents) ExtractHeaderLabels(labelCol int) []string {
	ll := make([]string, 0, 10)
	for _, re := range r {
		ll = append(ll, re.ExtractHeaderLabels(labelCol)...)
	}

	return ll
}

// Labelize converts labels into a row event.
func (r RowEvents) Labelize(cols []int, labelCol int, labels []string) RowEvents {
	out := make(RowEvents, 0, len(r))
	for _, re := range r {
		out = append(out, re.Labelize(cols, labelCol, labels))
	}

	return out
}

// Customize returns custom row events based on columns layout.
func (r RowEvents) Customize(cols []int) RowEvents {
	ee := make(RowEvents, 0, len(cols))
	for _, re := range r {
		ee = append(ee, re.Customize(cols))
	}
	return ee
}

// Diff returns true if the event changed.
func (r RowEvents) Diff(re RowEvents, ageCol int) bool {
	if len(r) != len(re) {
		return true
	}

	for i := range r {
		if r[i].Diff(re[i], ageCol) {
			return true
		}
	}

	return false
}

// Clone returns a rowevents deep copy.
func (r RowEvents) Clone() RowEvents {
	res := make(RowEvents, len(r))
	for i, re := range r {
		res[i] = re.Clone()
	}

	return res
}

// Upsert add or update a row if it exists.
func (r RowEvents) Upsert(re RowEvent) RowEvents {
	if idx, ok := r.FindIndex(re.Row.ID); ok {
		r[idx] = re
	} else {
		r = append(r, re)
	}
	return r
}

// Delete removes an element by id.
func (r RowEvents) Delete(id string) RowEvents {
	victim, ok := r.FindIndex(id)
	if !ok {
		return r
	}
	return append(r[0:victim], r[victim+1:]...)
}

// Clear delete all row events.
func (r RowEvents) Clear() RowEvents {
	return RowEvents{}
}

// FindIndex locates a row index by id. Returns false is not found.
func (r RowEvents) FindIndex(id string) (int, bool) {
	for i, re := range r {
		if re.Row.ID == id {
			return i, true
		}
	}

	return 0, false
}

// Sort rows based on column index and order.
func (r RowEvents) Sort(ns string, sortCol int, isDuration, numCol, isCapacity, asc bool) {
	if sortCol == -1 {
		return
	}

	t := RowEventSorter{
		NS:         ns,
		Events:     r,
		Index:      sortCol,
		Asc:        asc,
		IsNumber:   numCol,
		IsDuration: isDuration,
		IsCapacity: isCapacity,
	}
	sort.Sort(t)
}

// ----------------------------------------------------------------------------

// RowEventSorter sorts row events by a given colon.
type RowEventSorter struct {
	Events     RowEvents
	Index      int
	NS         string
	IsNumber   bool
	IsDuration bool
	IsCapacity bool
	Asc        bool
}

func (r RowEventSorter) Len() int {
	return len(r.Events)
}

func (r RowEventSorter) Swap(i, j int) {
	r.Events[i], r.Events[j] = r.Events[j], r.Events[i]
}

func (r RowEventSorter) Less(i, j int) bool {
	f1, f2 := r.Events[i].Row.Fields, r.Events[j].Row.Fields
	id1, id2 := r.Events[i].Row.ID, r.Events[j].Row.ID
	less := Less(r.IsNumber, r.IsDuration, r.IsCapacity, id1, id2, f1[r.Index], f2[r.Index])
	if r.Asc {
		return less
	}

	return !less
}

// ----------------------------------------------------------------------------

// // IdSorter sorts row events by a given id.
// type IdSorter struct {
// 	Ids    map[string]int
// 	Events RowEvents
// }

// func (s IdSorter) Len() int {
// 	return len(s.Events)
// }

// func (s IdSorter) Swap(i, j int) {
// 	s.Events[i], s.Events[j] = s.Events[j], s.Events[i]
// }

// func (s IdSorter) Less(i, j int) bool {
// 	return s.Ids[s.Events[i].Row.ID] < s.Ids[s.Events[j].Row.ID]
// }

// ----------------------------------------------------------------------------

// // StringSet represents a collection of unique strings.
// type StringSet []string

// // Add adds a new item in the set.
// func (ss StringSet) Add(item string) StringSet {
// 	if ss.In(item) {
// 		return ss
// 	}
// 	return append(ss, item)
// }

// // In checks if a string is in the set.
// func (ss StringSet) In(item string) bool {
// 	return ss.indexOf(item) >= 0
// }

// func (ss StringSet) indexOf(item string) int {
// 	for i, s := range ss {
// 		if s == item {
// 			return i
// 		}
// 	}
// 	return -1
// }
