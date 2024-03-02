// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import "reflect"

// DeltaRow represents a collection of row deltas between old and new row.
type DeltaRow []string

// NewDeltaRow computes the delta between 2 rows.
func NewDeltaRow(o, n Row, h Header) DeltaRow {
	deltas := make(DeltaRow, len(o.Fields))
	for i, old := range o.Fields {
		if old != "" && old != n.Fields[i] && !h.IsTimeCol(i) {
			deltas[i] = old
		}
	}

	return deltas
}

// Labelize returns a new deltaRow based on labels.
func (d DeltaRow) Labelize(cols []int, labelCol int) DeltaRow {
	if len(d) == 0 {
		return d
	}
	_, vals := sortLabels(labelize(d[labelCol]))
	out := make(DeltaRow, 0, len(cols)+len(vals))
	for _, i := range cols {
		out = append(out, d[i])
	}
	for _, v := range vals {
		out = append(out, v)
	}

	return out
}

// Diff returns true if deltas differ or false otherwise.
func (d DeltaRow) Diff(r DeltaRow, ageCol int) bool {
	if len(d) != len(r) {
		return true
	}

	if ageCol < 0 || ageCol >= len(d) {
		return !reflect.DeepEqual(d, r)
	}

	if !reflect.DeepEqual(d[:ageCol], r[:ageCol]) {
		return true
	}
	if ageCol+1 >= len(d) {
		return false
	}

	return !reflect.DeepEqual(d[ageCol+1:], r[ageCol+1:])
}

// Customize returns a subset of deltas.
func (d DeltaRow) Customize(cols []int, out DeltaRow) {
	if d.IsBlank() {
		return
	}
	for i, c := range cols {
		if c < 0 {
			continue
		}
		if c < len(d) && i < len(out) {
			out[i] = d[c]
		}
	}
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
	copy(res, d)

	return res
}
