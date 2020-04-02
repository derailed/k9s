package render

import (
	"reflect"
	"sort"
	"strconv"
	"time"

	"vbom.ml/util/sortorder"
)

// Fields represents a collection of row fields.
type Fields []string

// Customize returns a subset of fields.
func (f Fields) Customize(cols []int, out Fields) {
	for i, c := range cols {
		if c < 0 {
			out[i] = NAValue
			continue
		}
		if c < len(f) {
			out[i] = f[c]
		}
	}
}

// Diff returns true if fields differ or false otherwise.
func (f Fields) Diff(ff Fields, ageCol int) bool {
	if ageCol < 0 {
		return !reflect.DeepEqual(f[:len(f)-1], ff[:len(ff)-1])
	}
	if !reflect.DeepEqual(f[:ageCol], ff[:ageCol]) {
		return true
	}
	return !reflect.DeepEqual(f[ageCol+1:], ff[ageCol+1:])
}

// Clone returns a copy of the fields.
func (f Fields) Clone() Fields {
	cp := make(Fields, len(f))
	copy(cp, f)

	return cp
}

// ----------------------------------------------------------------------------

// Row represents a colllection of columns.
type Row struct {
	ID     string
	Fields Fields
}

// NewRow returns a new row with initialized fields.
func NewRow(size int) Row {
	return Row{Fields: make([]string, size)}
}

// Customize returns a row subset based on given col indices.
func (r Row) Customize(cols []int) Row {
	out := NewRow(len(cols))
	r.Fields.Customize(cols, out.Fields)
	out.ID = r.ID

	return out
}

// Diff returns true if row differ or false otherwise.
func (r Row) Diff(ro Row, ageCol int) bool {
	if r.ID != ro.ID {
		return true
	}
	return r.Fields.Diff(ro.Fields, ageCol)
}

// Clone copies a row.
func (r Row) Clone() Row {
	return Row{
		ID:     r.ID,
		Fields: r.Fields.Clone(),
	}
}

// Len returns the length of the row.
func (r Row) Len() int {
	return len(r.Fields)
}

// ----------------------------------------------------------------------------

// Rows represents a collection of rows.
type Rows []Row

// Delete removes an element by id.
func (rr Rows) Delete(id string) Rows {
	idx, ok := rr.Find(id)
	if !ok {
		return rr
	}

	if idx == 0 {
		return rr[1:]
	}
	if idx+1 == len(rr) {
		return rr[:len(rr)-1]
	}

	return append(rr[:idx], rr[idx+1:]...)
}

// Upsert adds a new item.
func (rr Rows) Upsert(r Row) Rows {
	idx, ok := rr.Find(r.ID)
	if !ok {
		return append(rr, r)
	}
	rr[idx] = r

	return rr
}

// Find locates a row by id. Retturns false is not found.
func (rr Rows) Find(id string) (int, bool) {
	for i, r := range rr {
		if r.ID == id {
			return i, true
		}
	}

	return 0, false
}

// Sort rows based on column index and order.
func (rr Rows) Sort(col int, asc bool) {
	t := RowSorter{Rows: rr, Index: col, Asc: asc}
	sort.Sort(t)
}

// ----------------------------------------------------------------------------

// RowSorter sorts rows.
type RowSorter struct {
	Rows  Rows
	Index int
	Asc   bool
}

func (s RowSorter) Len() int {
	return len(s.Rows)
}

func (s RowSorter) Swap(i, j int) {
	s.Rows[i], s.Rows[j] = s.Rows[j], s.Rows[i]
}

func (s RowSorter) Less(i, j int) bool {
	return Less(s.Asc, s.Rows[i].Fields[s.Index], s.Rows[j].Fields[s.Index])
}

// ----------------------------------------------------------------------------
// Helpers...

func toAgeDuration(dur string) string {
	d, err := time.ParseDuration(dur)
	if err != nil {
		return durationToSeconds(dur)
	}

	return strconv.Itoa(int(d.Seconds()))
}

// Less return true if c1 < c2.
func Less(asc bool, c1, c2 string) bool {
	c1, c2 = toAgeDuration(c1), toAgeDuration(c2)
	b := sortorder.NaturalLess(c1, c2)
	if asc {
		return b
	}
	return !b
}
