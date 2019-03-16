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
	c1 := s.rows[i][s.index]
	c2 := s.rows[j][s.index]

	if m1, ok := isMetric(c1); ok {
		m2, _ := isMetric(c2)
		i1, _ := strconv.Atoi(m1)
		i2, _ := strconv.Atoi(m2)
		if s.asc {
			return i1 < i2
		}
		return i1 > i2
	}

	c := strings.Compare(c1, c2)
	if s.asc {
		return c < 0
	}
	return c > 0
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
	c1 := s.rows[i]
	c2 := s.rows[j]

	if m1, ok := isMetric(c1); ok {
		m2, _ := isMetric(c2)
		i1, _ := strconv.Atoi(m1)
		i2, _ := strconv.Atoi(m2)
		if s.asc {
			return i1 < i2
		}
		return i1 > i2
	}

	c := strings.Compare(c1, c2)
	if s.asc {
		return c < 0
	}
	return c > 0
}

// ----------------------------------------------------------------------------
// Helpers...

var metricRX = regexp.MustCompile(`\A(\d+)(m|Mi)\z`)

func isMetric(s string) (string, bool) {
	if m := metricRX.FindStringSubmatch(s); len(m) == 3 {
		return m[1], true
	}
	return s, false
}
