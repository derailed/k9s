package views

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/resource"
)

type rowSorter struct {
	rows  resource.Rows
	index int
	asc   bool
}

func (s rowSorter) Len() int {
	return len(s.rows)
}

func (s rowSorter) Swap(i, j int) {
	s.rows[i], s.rows[j] = s.rows[j], s.rows[i]
}

func (s rowSorter) Less(i, j int) bool {
	return less(s.asc, s.rows[i][s.index], s.rows[j][s.index])
}

// ----------------------------------------------------------------------------

type groupSorter struct {
	rows []string
	asc  bool
}

func (s groupSorter) Len() int {
	return len(s.rows)
}

func (s groupSorter) Swap(i, j int) {
	s.rows[i], s.rows[j] = s.rows[j], s.rows[i]
}

func (s groupSorter) Less(i, j int) bool {
	return less(s.asc, s.rows[i], s.rows[j])
}

// ----------------------------------------------------------------------------
// Helpers...

func less(asc bool, c1, c2 string) bool {
	if o, ok := isMetricSort(asc, c1, c2); ok {
		return o
	}

	if o, ok := isIntegerSort(asc, c1, c2); ok {
		return o
	}

	c := strings.Compare(c1, c2)
	if asc {
		return c < 0
	}
	return c > 0
}

func isMetricSort(asc bool, c1, c2 string) (bool, bool) {
	m1, ok := isMetric(c1)
	if !ok {
		return false, false
	}
	m2, _ := isMetric(c2)
	i1, _ := strconv.Atoi(m1)
	i2, _ := strconv.Atoi(m2)
	if asc {
		return i1 < i2, true
	}
	return i1 > i2, true
}

func isIntegerSort(asc bool, c1, c2 string) (bool, bool) {
	n1, err := strconv.Atoi(c1)
	if err != nil {
		return false, false
	}
	n2, _ := strconv.Atoi(c2)
	if asc {
		return n1 < n2, true
	}
	return n1 > n2, true
}

var metricRX = regexp.MustCompile(`\A(\d+)(m|Mi)\z`)

func isMetric(s string) (string, bool) {
	if m := metricRX.FindStringSubmatch(s); len(m) == 3 {
		return m[1], true
	}
	return s, false
}
