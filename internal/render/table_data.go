package render

import (
	"sync"
)

// TableData tracks a K8s resource for tabular display.
type TableData struct {
	Header    HeaderRow
	RowEvents RowEvents
	Namespace string
	Mutex     *sync.RWMutex
}

// NewTableData returns a new table.
func NewTableData() *TableData {
	return &TableData{Mutex: &sync.RWMutex{}}
}

// Clear clears out the entire table.
func (t *TableData) Clear() {
	t.Header, t.RowEvents = t.Header.Clear(), t.RowEvents.Clear()
}

// Clone returns a copy of the table
func (t *TableData) Clone() TableData {
	return cloneTable(*t)
}

func cloneTable(t TableData) TableData {
	return t
}

// Update computes row deltas and update the table data.
func (t *TableData) Update(rows Rows) {
	empty := len(t.RowEvents) == 0
	kk := make([]string, 0, len(rows))
	var blankDelta DeltaRow
	for _, row := range rows {
		kk = append(kk, row.ID)
		if empty {
			t.RowEvents = append(t.RowEvents, NewRowEvent(EventAdd, row))
			continue
		}

		if index, ok := t.RowEvents.FindIndex(row.ID); ok {
			delta := NewDeltaRow(t.RowEvents[index].Row, row, t.Header.HasAge())
			if delta.IsBlank() {
				t.RowEvents[index].Kind, t.RowEvents[index].Deltas = EventUnchanged, blankDelta
				t.RowEvents[index].Row = row
			} else {
				t.RowEvents[index] = NewDeltaRowEvent(row, delta)
			}
			continue
		}
		t.RowEvents = append(t.RowEvents, NewRowEvent(EventAdd, row))
	}

	if !empty {
		t.Delete(kk)
	}
}

// Delete delete items in cache that are no longer valid.
func (t *TableData) Delete(newKeys []string) {
	var victims []string
	for _, re := range t.RowEvents {
		var found bool
		for i, key := range newKeys {
			if key == re.Row.ID {
				found = true
				newKeys = append(newKeys[:i], newKeys[i+1:]...)
				break
			}
		}
		if !found {
			victims = append(victims, re.Row.ID)
		}
	}

	for _, id := range victims {
		t.RowEvents = t.RowEvents.Delete(id)
	}
}

// Diff checks if two tables are equal.
func (t *TableData) Diff(table TableData) bool {
	if t.Namespace != table.Namespace {
		return true
	}
	if t.Header.Changed(table.Header) {
		return true
	}
	if t.RowEvents.Changed(table.RowEvents) {
		return true
	}

	return false
}
