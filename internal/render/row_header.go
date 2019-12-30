package render

import "reflect"

const ageCol = "AGE"

// Header represent a table header
type Header struct {
	Name      string
	Align     int
	Decorator DecoratorFunc
}

// Clone copies a header.
func (h Header) Clone() Header {
	return h
}

// ----------------------------------------------------------------------------

// HeaderRow represents a table header.
type HeaderRow []Header

func (hh HeaderRow) Clone() HeaderRow {
	h := make(HeaderRow, len(hh))
	for i, v := range hh {
		h[i] = v.Clone()
	}

	return h
}

// Clear clears out the header row.
func (hh HeaderRow) Clear() HeaderRow {
	return HeaderRow{}
}

// Changed returns true if the header changed.
func (hh HeaderRow) Changed(h HeaderRow) bool {
	if len(hh) != len(h) {
		return true
	}
	return !reflect.DeepEqual(hh.Columns(), h.Columns())
}

// Columns return header  as a collection of strings.
func (h HeaderRow) Columns() []string {
	cc := make([]string, len(h))
	for i, c := range h {
		cc[i] = c.Name
	}

	return cc
}

// HasAge returns true if table has an age column.
func (h HeaderRow) HasAge() bool {
	for _, r := range h {
		if r.Name == ageCol {
			return true
		}
	}

	return false
}

// AgeCol checks if given column index is the age column.
func (h HeaderRow) AgeCol(col int) bool {
	if !h.HasAge() {
		return false
	}
	return col == len(h)-1
}
