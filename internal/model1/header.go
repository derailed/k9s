// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import (
	"fmt"
	"log/slog"
	"reflect"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/slogs"
	"k8s.io/apimachinery/pkg/util/sets"
)

const ageCol = "AGE"

type Attrs struct {
	Align     int
	Decorator DecoratorFunc
	Wide      bool
	Show      bool
	MX        bool
	MXC, MXM  bool
	Time      bool
	Capacity  bool
	VS        bool
	Hide      bool
}

func (a Attrs) Merge(b Attrs) Attrs {
	a.MX = b.MX
	a.MXC = b.MXC
	a.MXM = b.MXM
	a.Decorator = b.Decorator
	a.VS = b.VS

	if a.Align == 0 {
		a.Align = b.Align
	}

	if !a.Hide {
		a.Hide = b.Hide
	}
	if !a.Show && !a.Wide {
		a.Wide = b.Wide
	}

	if !a.Time {
		a.Time = b.Time
	}
	if !a.Capacity {
		a.Capacity = b.Capacity
	}

	return a
}

// HeaderColumn represent a table header.
type HeaderColumn struct {
	Attrs
	Name string
}

func (h HeaderColumn) String() string {
	return fmt.Sprintf("%s [%d::%t::%t::%t]", h.Name, h.Align, h.Wide, h.MX, h.Time)
}

// Clone copies a header.
func (h HeaderColumn) Clone() HeaderColumn {
	return h
}

// ----------------------------------------------------------------------------

// Header represents a table header.
type Header []HeaderColumn

func (h Header) Clear() Header {
	h = h[:0]

	return h
}

// Clone duplicates a header.
func (h Header) Clone() Header {
	he := make(Header, 0, len(h))
	for _, h := range h {
		he = append(he, h.Clone())
	}

	return he
}

// Labelize returns a new Header based on labels.
func (h Header) Labelize(cols []int, labelCol int, rr *RowEvents) Header {
	header := make(Header, 0, len(cols)+1)
	for _, c := range cols {
		header = append(header, h[c])
	}
	cc := rr.ExtractHeaderLabels(labelCol)
	for _, c := range cc {
		header = append(header, HeaderColumn{Name: c})
	}

	return header
}

// MapIndices returns a collection of mapped column indices based of the requested columns.
func (h Header) MapIndices(cols []string, wide bool) []int {
	ii := make([]int, 0, len(cols))
	cc := make(map[int]struct{}, len(cols))
	for _, col := range cols {
		idx, ok := h.IndexOf(col, true)
		if !ok {
			slog.Warn("Column not found on resource", slogs.ColName, col)
		}
		ii, cc[idx] = append(ii, idx), struct{}{}
	}
	if !wide {
		return ii
	}

	for i := range h {
		if _, ok := cc[i]; ok {
			continue
		}
		ii = append(ii, i)
	}
	return ii
}

// Customize builds a header from custom col definitions.
func (h Header) Customize(cols []string, wide bool) Header {
	if len(cols) == 0 {
		return h
	}
	cc := make(Header, 0, len(h))
	xx := make(map[int]struct{}, len(h))
	for _, c := range cols {
		idx, ok := h.IndexOf(c, true)
		if !ok {
			slog.Warn("Column is not available on this resource", slogs.ColName, c)
			cc = append(cc, HeaderColumn{Name: c})
			continue
		}
		xx[idx] = struct{}{}
		col := h[idx].Clone()
		col.Wide = false
		cc = append(cc, col)
	}
	if !wide {
		return cc
	}

	for i, c := range h {
		if _, ok := xx[i]; ok {
			continue
		}
		col := c.Clone()
		col.Wide = true
		cc = append(cc, col)
	}

	return cc
}

// Diff returns true if the header changed.
func (h Header) Diff(header Header) bool {
	if len(h) != len(header) {
		return true
	}
	return !reflect.DeepEqual(h, header)
}

// FilterColIndices return viewable col header indices.
func (h Header) FilterColIndices(ns string, wide bool) sets.Set[int] {
	if len(h) == 0 {
		return nil
	}
	nsed := client.IsNamespaced(ns)

	cc := sets.New[int]()
	for i, c := range h {
		if c.Name == "AGE" || !wide && c.Wide || c.Hide || (nsed && c.Name == "NAMESPACE") {
			continue
		}
		cc.Insert(i)
	}

	return cc
}

// ColumnNames return header col names
func (h Header) ColumnNames(wide bool) []string {
	if len(h) == 0 {
		return nil
	}
	cc := make([]string, 0, len(h))
	for _, c := range h {
		if !wide && c.Wide {
			continue
		}
		cc = append(cc, c.Name)
	}

	return cc
}

// HasAge returns true if table has an age column.
func (h Header) HasAge() bool {
	_, ok := h.IndexOf(ageCol, true)

	return ok
}

// IsMetricsCol checks if given column index represents metrics.
func (h Header) IsMetricsCol(col int) bool {
	if col < 0 || col >= len(h) {
		return false
	}

	return h[col].MX
}

// IsTimeCol checks if given column index represents a timestamp.
func (h Header) IsTimeCol(col int) bool {
	if col < 0 || col >= len(h) {
		return false
	}

	return h[col].Time
}

// IsCapacityCol checks if given column index represents a capacity.
func (h Header) IsCapacityCol(col int) bool {
	if col < 0 || col >= len(h) {
		return false
	}

	return h[col].Capacity
}

// IndexOf returns the col index or -1 if none.
func (h Header) IndexOf(colName string, includeWide bool) (int, bool) {
	for i, c := range h {
		if c.Wide && !includeWide {
			continue
		}
		if c.Name == colName {
			return i, true
		}
	}
	return -1, false
}

// Dump for debugging.
func (h Header) Dump() {
	slog.Debug("HEADER")
	for i, c := range h {
		slog.Debug(fmt.Sprintf("%d %q -- %t", i, c.Name, c.Wide))
	}
}
