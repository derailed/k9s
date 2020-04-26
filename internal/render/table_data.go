package render

import "github.com/derailed/k9s/internal/client"

// TableData tracks a K8s resource for tabular display.
type TableData struct {
	Header    Header
	RowEvents RowEvents
	Namespace string
}

// NewTableData returns a new table.
func NewTableData() *TableData {
	return &TableData{}
}

// Labelize prints out specific label columns
func (t *TableData) Labelize(labels []string) TableData {
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

	return data
}

// Customize returns a new model with customized column layout.
func (t *TableData) Customize(cols []string, wide bool) TableData {
	res := TableData{
		Namespace: t.Namespace,
		Header:    t.Header.Customize(cols, wide),
	}
	ids := t.Header.MapIndices(cols, wide)
	res.RowEvents = t.RowEvents.Customize(ids)

	return res
}

// Clear clears out the entire table.
func (t *TableData) Clear() {
	t.Header, t.RowEvents = Header{}, RowEvents{}
}

// Clone returns a copy of the table
func (t *TableData) Clone() TableData {
	return TableData{
		Header:    t.Header.Clone(),
		RowEvents: t.RowEvents.Clone(),
		Namespace: t.Namespace,
	}
}

// SetHeader sets table header.
func (t *TableData) SetHeader(ns string, h Header) {
	t.Namespace, t.Header = ns, h
}

// Update computes row deltas and update the table data.
func (t *TableData) Update(rows Rows) {
	empty := len(t.RowEvents) == 0
	kk := make(map[string]struct{}, len(rows))
	var blankDelta DeltaRow
	for _, row := range rows {
		kk[row.ID] = struct{}{}
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
				t.RowEvents[index] = NewRowEventWithDeltas(row, delta)
			}
			continue
		}
		t.RowEvents = append(t.RowEvents, NewRowEvent(EventAdd, row))
	}

	if !empty {
		t.Delete(kk)
	}
}

// Delete removes items in cache that are no longer valid.
func (t *TableData) Delete(newKeys map[string]struct{}) {
	var victims []string
	for _, re := range t.RowEvents {
		if _, ok := newKeys[re.Row.ID]; !ok {
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

	if t.Header.Diff(table.Header) {
		return true
	}

	return t.RowEvents.Diff(table.RowEvents, t.Header.IndexOf("AGE", true))
}
