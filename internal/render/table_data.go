// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"sync"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
)

// TableData tracks a K8s resource for tabular display.
type TableData struct {
	Header    Header
	RowEvents *RowEvents
	Namespace string
	mx        sync.RWMutex
}

// NewTableData returns a new table.
func NewTableData() *TableData {
	return &TableData{
		RowEvents: NewRowEvents(10),
	}
}

// Empty checks if there are no entries.
func (t *TableData) Empty() bool {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return t.RowEvents.Empty()
}

// Count returns the number of entries.
func (t *TableData) Count() int {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return t.RowEvents.Len()
}

// IndexOfHeader return the index of the header.
func (t *TableData) IndexOfHeader(h string) int {
	return t.Header.IndexOf(h, false)
}

// Labelize prints out specific label columns.
func (t *TableData) Labelize(labels []string) *TableData {
	labelCol := t.Header.IndexOf("LABELS", true)
	cols := []int{0, 1}
	if client.IsNamespaced(t.Namespace) {
		cols = cols[1:]
	}
	data := TableData{
		Namespace: t.Namespace,
		Header:    t.Header.Labelize(cols, labelCol, t.RowEvents),
	}
	data.RowEvents = t.RowEvents.Labelize(cols, labelCol, labels)

	return &data
}

// Customize returns a new model with customized column layout.
func (t *TableData) Customize(cols []string, wide bool) *TableData {
	res := TableData{
		Namespace: t.Namespace,
		Header:    t.Header.Customize(cols, wide),
	}
	ids := t.Header.MapIndices(cols, wide)
	res.RowEvents = t.RowEvents.Customize(ids)

	return &res
}

// Clear clears out the entire table.
func (t *TableData) Clear() {
	t.mx.Lock()
	defer t.mx.Unlock()

	t.Header = t.Header.Clear()
	t.RowEvents.Clear()
}

// Clone returns a copy of the table.
func (t *TableData) Clone() *TableData {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return &TableData{
		Header:    t.Header.Clone(),
		RowEvents: t.RowEvents.Clone(),
		Namespace: t.Namespace,
	}
}

func (t *TableData) GetHeader() Header {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return t.Header
}

func (t *TableData) ColumnNames(w bool) []string {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return t.Header.ColumnNames(w)
}

// SetHeader sets table header.
func (t *TableData) SetHeader(ns string, h Header) {
	t.mx.Lock()
	defer t.mx.Unlock()

	t.Namespace, t.Header = ns, h
}

// Update computes row deltas and update the table data.
func (t *TableData) Update(rows Rows) {
	defer func(ti time.Time) {
		log.Debug().Msgf("  TDA-UPDATE  [%d] (%s)", len(rows), time.Since(ti))
	}(time.Now())

	empty := t.Empty()
	kk := make(map[string]struct{}, len(rows))
	var blankDelta DeltaRow
	t.mx.Lock()
	{
		for _, row := range rows {
			kk[row.ID] = struct{}{}
			if empty {
				t.RowEvents.Add(NewRowEvent(EventAdd, row))
				continue
			}
			if index, ok := t.RowEvents.FindIndex(row.ID); ok {
				ev, ok := t.RowEvents.At(index)
				if !ok {
					continue
				}
				delta := NewDeltaRow(ev.Row, row, t.Header)
				if delta.IsBlank() {
					ev.Kind, ev.Deltas, ev.Row = EventUnchanged, blankDelta, row
					t.RowEvents.Set(index, ev)
				} else {
					t.RowEvents.Set(index, NewRowEventWithDeltas(row, delta))
				}
				continue
			}
			t.RowEvents.Add(NewRowEvent(EventAdd, row))
		}
	}
	t.mx.Unlock()

	if !empty {
		t.Delete(kk)
	}
}

// Delete removes items in cache that are no longer valid.
func (t *TableData) Delete(newKeys map[string]struct{}) {
	t.mx.Lock()
	{
		victims := make([]string, 0, 10)
		t.RowEvents.Range(func(_ int, e RowEvent) bool {
			if _, ok := newKeys[e.Row.ID]; !ok {
				victims = append(victims, e.Row.ID)
			} else {
				delete(newKeys, e.Row.ID)
			}
			return true
		})
		for _, id := range victims {
			t.RowEvents.Delete(id)
		}
	}
	t.mx.Unlock()
}

// Diff checks if two tables are equal.
func (t *TableData) Diff(t2 *TableData) bool {
	if t2 == nil || t.Namespace != t2.Namespace || t.Header.Diff(t2.Header) {
		return true
	}

	return t.RowEvents.Diff(t2.RowEvents, t.Header.IndexOf("AGE", true))
}
