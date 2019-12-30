package render

// DeltaRow represents a collection of row detlas between old and new row.
type DeltaRow []string

// NewDeltaRow computes the delta between 2 rows.
func NewDeltaRow(o, n Row, excludeLast bool) DeltaRow {
	deltas := make(DeltaRow, len(o.Fields))
	// Exclude age col
	oldFields := o.Fields[:len(o.Fields)-1]
	if !excludeLast {
		oldFields = o.Fields[:len(o.Fields)]
	}
	for i, old := range oldFields {
		if old != "" && old != n.Fields[i] {
			deltas[i] = old
		}
	}

	return deltas
}

// IsBlank asserts a row has no values in it.
func (d DeltaRow) IsBlank() bool {
	if len(d) == 0 {
		return true
	}

	for _, v := range d {
		if v != "" {
			return false
		}
	}

	return true
}

// Clone returns a delta copy.
func (d DeltaRow) Clone() DeltaRow {
	res := make(DeltaRow, len(d))
	for i, f := range d {
		res[i] = f
	}

	return res
}
