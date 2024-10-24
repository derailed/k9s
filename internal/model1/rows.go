// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import "sort"

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

// Find locates a row by id. Returns false is not found.
func (rr Rows) Find(id string) (int, bool) {
	for i, r := range rr {
		if r.ID == id {
			return i, true
		}
	}

	return 0, false
}

// Sort rows based on column index and order.
func (rr Rows) Sort(col int, asc, isNum, isDur, isCapacity bool) {
	t := RowSorter{
		Rows:       rr,
		Index:      col,
		IsNumber:   isNum,
		IsDuration: isDur,
		IsCapacity: isCapacity,
		Asc:        asc,
	}
	sort.Sort(t)
}
