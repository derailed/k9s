// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import (
	"fmt"
	"sort"
)

type ReRangeFn func(int, RowEvent) bool

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

type reIndex map[string]int

// RowEvents a collection of row events.
type RowEvents struct {
	events []RowEvent
	index  reIndex
}

func NewRowEvents(size int) *RowEvents {
	return &RowEvents{
		events: make([]RowEvent, 0, size),
		index:  make(reIndex, size),
	}
}

func NewRowEventsWithEvts(ee ...RowEvent) *RowEvents {
	re := NewRowEvents(len(ee))
	for _, e := range ee {
		re.Add(e)
	}

	return re
}

func (r *RowEvents) reindex() {
	for i, e := range r.events {
		r.index[e.Row.ID] = i
	}
}

func (r *RowEvents) At(i int) (RowEvent, bool) {
	if i < 0 || i > len(r.events) {
		return RowEvent{}, false
	}

	return r.events[i], true
}

func (r *RowEvents) Set(i int, re RowEvent) {
	r.events[i] = re
	r.index[re.Row.ID] = i
}

func (r *RowEvents) Add(re RowEvent) {
	r.events = append(r.events, re)
	r.index[re.Row.ID] = len(r.events) - 1
}

// ExtractHeaderLabels extract header labels.
func (r *RowEvents) ExtractHeaderLabels(labelCol int) []string {
	ll := make([]string, 0, 10)
	for _, re := range r.events {
		ll = append(ll, re.ExtractHeaderLabels(labelCol)...)
	}

	return ll
}

// Labelize converts labels into a row event.
func (r *RowEvents) Labelize(cols []int, labelCol int, labels []string) *RowEvents {
	out := make([]RowEvent, 0, len(r.events))
	for _, re := range r.events {
		out = append(out, re.Labelize(cols, labelCol, labels))
	}

	return NewRowEventsWithEvts(out...)
}

// Customize returns custom row events based on columns layout.
func (r *RowEvents) Customize(cols []int) *RowEvents {
	ee := make([]RowEvent, 0, len(cols))
	for _, re := range r.events {
		ee = append(ee, re.Customize(cols))
	}

	return NewRowEventsWithEvts(ee...)
}

// Diff returns true if the event changed.
func (r *RowEvents) Diff(re *RowEvents, ageCol int) bool {
	if len(r.events) != len(re.events) {
		return true
	}
	for i := range r.events {
		if r.events[i].Diff(re.events[i], ageCol) {
			return true
		}
	}

	return false
}

// Clone returns a deep copy.
func (r *RowEvents) Clone() *RowEvents {
	re := make([]RowEvent, 0, len(r.events))
	for _, e := range r.events {
		re = append(re, e.Clone())
	}

	return NewRowEventsWithEvts(re...)
}

// Upsert add or update a row if it exists.
func (r *RowEvents) Upsert(re RowEvent) {
	if idx, ok := r.FindIndex(re.Row.ID); ok {
		r.events[idx] = re
	} else {
		r.Add(re)
	}
}

// Delete removes an element by id.
func (r *RowEvents) Delete(fqn string) error {
	victim, ok := r.FindIndex(fqn)
	if !ok {
		return fmt.Errorf("unable to delete row with fqn: %q", fqn)
	}
	r.events = append(r.events[0:victim], r.events[victim+1:]...)
	delete(r.index, fqn)
	r.reindex()

	return nil
}

func (r *RowEvents) Len() int {
	return len(r.events)
}

func (r *RowEvents) Empty() bool {
	return len(r.events) == 0
}

// Clear delete all row events.
func (r *RowEvents) Clear() {
	r.events = r.events[:0]
	for k := range r.index {
		delete(r.index, k)
	}
}

func (r *RowEvents) Range(f ReRangeFn) {
	for i, e := range r.events {
		if !f(i, e) {
			return
		}
	}
}

func (r *RowEvents) Get(id string) (RowEvent, bool) {
	i, ok := r.index[id]
	if !ok {
		return RowEvent{}, false
	}

	return r.At(i)
}

// FindIndex locates a row index by id. Returns false is not found.
func (r *RowEvents) FindIndex(id string) (int, bool) {
	i, ok := r.index[id]

	return i, ok
}

// Sort rows based on column index and order.
func (r *RowEvents) Sort(ns string, sortCol int, isDuration, numCol, isCapacity, asc bool) {
	if sortCol == -1 || r == nil {
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
	r.reindex()
}

// ----------------------------------------------------------------------------

// RowEventSorter sorts row events by a given colon.
type RowEventSorter struct {
	Events     *RowEvents
	Index      int
	NS         string
	IsNumber   bool
	IsDuration bool
	IsCapacity bool
	Asc        bool
}

func (r RowEventSorter) Len() int {
	return len(r.Events.events)
}

func (r RowEventSorter) Swap(i, j int) {

	r.Events.events[i], r.Events.events[j] = r.Events.events[j], r.Events.events[i]
}

func (r RowEventSorter) Less(i, j int) bool {
	f1, f2 := r.Events.events[i].Row.Fields, r.Events.events[j].Row.Fields
	id1, id2 := r.Events.events[i].Row.ID, r.Events.events[j].Row.ID
	less := Less(r.IsNumber, r.IsDuration, r.IsCapacity, id1, id2, f1[r.Index], f2[r.Index])
	if r.Asc {
		return less
	}

	return !less
}
