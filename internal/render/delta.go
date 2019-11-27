package render

// DeltaRow represents a collection of row detlas between old and new row.
type DeltaRow []string

// NewDeltaRow computes the delta between 2 rows.
func NewDeltaRow(o, n Row) DeltaRow {
	deltas := make(DeltaRow, len(o.Fields))
	// Exclude age col
	fields := o.Fields[:len(o.Fields)-1]
	for i, v := range fields {
		if v != "" && n.Fields[i] != v {
			deltas[i] = v
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
