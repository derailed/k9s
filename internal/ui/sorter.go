package ui

import (
	"strconv"
	"time"

	"github.com/derailed/k9s/internal/resource"
	res "k8s.io/apimachinery/pkg/api/resource"
	"vbom.ml/util/sortorder"
)

type (
	// SortFn represent a function that can sort columnar data.
	SortFn func(rows resource.Rows, sortCol SortColumn)

	// SortColumn represents a sortable column.
	SortColumn struct {
		index    int
		colCount int
		asc      bool
	}

	// RowSorter sorts rows.
	RowSorter struct {
		rows  resource.Rows
		index int
		asc   bool
	}
)

func (s RowSorter) Len() int {
	return len(s.rows)
}

func (s RowSorter) Swap(i, j int) {
	s.rows[i], s.rows[j] = s.rows[j], s.rows[i]
}

func (s RowSorter) Less(i, j int) bool {
	return less(s.asc, s.rows[i][s.index], s.rows[j][s.index])
}

// ----------------------------------------------------------------------------

// GroupSorter sorts a collection of rows.
type GroupSorter struct {
	rows []string
	asc  bool
}

func (s GroupSorter) Len() int {
	return len(s.rows)
}

func (s GroupSorter) Swap(i, j int) {
	s.rows[i], s.rows[j] = s.rows[j], s.rows[i]
}

func (s GroupSorter) Less(i, j int) bool {
	return less(s.asc, s.rows[i], s.rows[j])
}

// ----------------------------------------------------------------------------
// Helpers...

func less(asc bool, c1, c2 string) bool {
	if c1 == resource.NAValue && c2 != resource.NAValue {
		return false
	}
	if c1 != resource.NAValue && c2 == resource.NAValue {
		return true
	}

	if o, ok := isIntegerSort(asc, c1, c2); ok {
		return o
	}

	if o, ok := isMetricSort(asc, c1, c2); ok {
		return o
	}

	if o, ok := isDurationSort(asc, c1, c2); ok {
		return o
	}

	b := sortorder.NaturalLess(c1, c2)
	if asc {
		return b
	}
	return !b
}

func isDurationSort(asc bool, s1, s2 string) (bool, bool) {
	d1, ok1 := isDuration(s1)
	d2, ok2 := isDuration(s2)
	if !ok1 || !ok2 {
		return false, false
	}

	if asc {
		return d1 <= d2, true
	}
	return d1 >= d2, true
}

func isMetricSort(asc bool, c1, c2 string) (bool, bool) {
	q1, err1 := res.ParseQuantity(c1)
	q2, err2 := res.ParseQuantity(c2)
	if err1 != nil || err2 != nil {
		return false, false
	}

	if asc {
		return q1.Cmp(q2) <= 0, true
	}
	return q1.Cmp(q2) > 0, true
}

func isIntegerSort(asc bool, c1, c2 string) (bool, bool) {
	n1, err1 := strconv.Atoi(c1)
	n2, err2 := strconv.Atoi(c2)
	if err1 != nil || err2 != nil {
		return false, false
	}

	if asc {
		return n1 <= n2, true
	}
	return n1 > n2, true
}

func isDuration(s string) (time.Duration, bool) {
	d, err := time.ParseDuration(s)
	if err != nil {
		return d, false
	}
	return d, true
}
