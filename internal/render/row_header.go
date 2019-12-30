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

// Clone duplicates a header.
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
func (hh HeaderRow) Columns() []string {
	cc := make([]string, len(hh))
	for i, c := range hh {
		cc[i] = c.Name
	}

	return cc
}

// HasAge returns true if table has an age column.
func (hh HeaderRow) HasAge() bool {
	for _, r := range hh {
		if r.Name == ageCol {
			return true
		}
	}

	return false
}

// AgeCol checks if given column index is the age column.
func (hh HeaderRow) AgeCol(col int) bool {
	if !hh.HasAge() {
		return false
	}
	return col == len(hh)-1
}
