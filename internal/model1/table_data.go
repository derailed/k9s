// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/rs/zerolog/log"
	"github.com/sahilm/fuzzy"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

type (
	// SortFn represent a function that can sort columnar data.
	SortFn func(rows Rows, sortCol SortColumn)

	// SortColumn represents a sortable column.
	SortColumn struct {
		Name string
		ASC  bool
	}
)

const spacer = " "

type FilterOpts struct {
	Toast  bool
	Filter string
	Invert bool
}

// TableData tracks a K8s resource for tabular display.
type TableData struct {
	header    Header
	rowEvents *RowEvents
	namespace string
	gvr       client.GVR
	mx        sync.RWMutex
}

// NewTableData returns a new table.
func NewTableData(gvr client.GVR) *TableData {
	return &TableData{
		gvr:       gvr,
		rowEvents: NewRowEvents(10),
	}
}

func NewTableDataFull(gvr client.GVR, ns string, h Header, re *RowEvents) *TableData {
	t := NewTableDataWithRows(gvr, h, re)
	t.namespace = ns

	return t
}

func NewTableDataWithRows(gvr client.GVR, h Header, re *RowEvents) *TableData {
	t := NewTableData(gvr)
	t.header, t.rowEvents = h, re

	return t
}

func NewTableDataFromTable(td *TableData) *TableData {
	t := NewTableData(td.gvr)
	t.header = td.header
	t.rowEvents = td.rowEvents
	t.namespace = td.namespace

	return t
}

func (t *TableData) AddRow(re RowEvent) {
	t.rowEvents.Add(re)
}

func (t *TableData) SetRow(idx int, re RowEvent) {
	t.rowEvents.Set(idx, re)
}

func (t *TableData) FindRow(id string) (RowEvent, bool) {
	return t.rowEvents.Get(id)
}

func (t *TableData) RowAt(idx int) (RowEvent, bool) {
	return t.rowEvents.At(idx)
}

func (t *TableData) RowsRange(f ReRangeFn) {
	t.rowEvents.Range(f)
}

func (t *TableData) Sort(sc SortColumn) {
	col, idx := t.HeadCol(sc.Name, false)
	if idx < 0 {
		return
	}
	t.rowEvents.Sort(
		t.GetNamespace(),
		idx,
		col.Time,
		col.MX,
		col.Capacity,
		sc.ASC,
	)
}

func (t *TableData) Header() Header {
	return t.header
}

// HeaderCount returns the number of header cols.
func (t *TableData) HeaderCount() int {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return len(t.header)
}

func (t *TableData) HeadCol(n string, w bool) (HeaderColumn, int) {
	idx, ok := t.header.IndexOf(n, w)
	if !ok {
		return HeaderColumn{}, -1
	}

	return t.header[idx], idx
}

func (t *TableData) Filter(f FilterOpts) *TableData {
	td := NewTableDataFromTable(t)

	if f.Toast {
		td.rowEvents = t.filterToast()
	}
	if f.Filter == "" || internal.IsLabelSelector(f.Filter) {
		return td
	}
	if f, ok := internal.IsFuzzySelector(f.Filter); ok {
		td.rowEvents = t.fuzzyFilter(f)
		return td
	}
	rr, err := t.rxFilter(f.Filter, internal.IsInverseSelector(f.Filter))
	if err == nil {
		td.rowEvents = rr
	} else {
		log.Error().Err(err).Msg("rx filter failed")
	}

	return td
}

func (t *TableData) rxFilter(q string, inverse bool) (*RowEvents, error) {
	if inverse {
		q = q[1:]
	}
	rx, err := regexp.Compile(`(?i)(` + q + `)`)
	if err != nil {
		return nil, fmt.Errorf("invalid rx filter %q: %w", q, err)
	}

	var startIndex int
	if _, ok := t.header.IndexOf("NAMESPACE", true); ok && client.IsNamespaced(t.namespace) {
		startIndex = 1
	}
	rr := NewRowEvents(t.RowCount() / 2)
	ageIndex, _ := t.header.IndexOf("AGE", true)
	t.rowEvents.Range(func(_ int, re RowEvent) bool {
		ff := re.Row.Fields[startIndex:]
		if ageIndex >= 0 && ageIndex+1 <= len(ff) {
			ff = append(ff[0:ageIndex], ff[ageIndex+1:]...)
		}
		match := rx.MatchString(strings.Join(ff, spacer))
		if (inverse && !match) || (!inverse && match) {
			rr.Add(re)
		}

		return true
	})

	return rr, nil
}

func (t *TableData) fuzzyFilter(q string) *RowEvents {
	q = strings.TrimSpace(q)
	ss := make([]string, 0, t.RowCount()/2)
	t.rowEvents.Range(func(_ int, re RowEvent) bool {
		ss = append(ss, re.Row.ID)
		return true
	})

	mm := fuzzy.Find(q, ss)
	rr := NewRowEvents(t.RowCount() / 2)
	for _, m := range mm {
		re, ok := t.rowEvents.At(m.Index)
		if !ok {
			log.Error().Msgf("unable to find event for index in fuzzfilter: %d", m.Index)
			continue
		}
		rr.Add(re)
	}

	return rr
}

func (t *TableData) filterToast() *RowEvents {
	rr := NewRowEvents(10)
	idx, ok := t.header.IndexOf("VALID", true)
	if !ok {
		return rr
	}
	t.rowEvents.Range(func(_ int, re RowEvent) bool {
		if re.Row.Fields[idx] != "" {
			rr.Add(re)
		}
		return true
	})

	return rr
}

func (t *TableData) GetNamespace() string {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return t.namespace
}

func (t *TableData) Reset(ns string) {
	t.mx.Lock()
	{
		t.namespace = ns
	}
	t.mx.Unlock()

	t.Clear()
}

func (t *TableData) Reconcile(ctx context.Context, r Renderer, oo []runtime.Object) error {
	var rows Rows
	if len(oo) > 0 {
		if r.IsGeneric() {
			table, ok := oo[0].(*metav1.Table)
			if !ok {
				return fmt.Errorf("expecting a meta table but got %T", oo[0])
			}
			rows = make(Rows, len(table.Rows))
			if err := GenericHydrate(t.namespace, table, rows, r); err != nil {
				return err
			}
		} else {
			rows = make(Rows, len(oo))
			if err := Hydrate(t.namespace, oo, rows, r); err != nil {
				return err
			}
		}
	}

	t.Update(rows)
	t.SetHeader(t.namespace, r.Header(t.namespace))
	if t.HeaderCount() == 0 {
		return fmt.Errorf("fail to list resource %s", t.gvr)
	}

	return nil
}

// Empty checks if there are no entries.
func (t *TableData) Empty() bool {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return t.rowEvents.Empty()
}

func (t *TableData) SetRowEvents(re *RowEvents) {
	t.rowEvents = re
}

func (t *TableData) GetRowEvents() *RowEvents {
	return t.rowEvents
}

// RowCount returns the number of rows.
func (t *TableData) RowCount() int {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return t.rowEvents.Len()
}

// IndexOfHeader return the index of the header.
func (t *TableData) IndexOfHeader(h string) (int, bool) {
	return t.header.IndexOf(h, false)
}

// Labelize prints out specific label columns.
func (t *TableData) Labelize(labels []string) *TableData {
	idx, ok := t.header.IndexOf("LABELS", true)
	if !ok {
		return t
	}
	cols := []int{0, 1}
	if client.IsNamespaced(t.namespace) {
		cols = cols[1:]
	}
	data := TableData{
		namespace: t.namespace,
		header:    t.header.Labelize(cols, idx, t.rowEvents),
	}
	data.rowEvents = t.rowEvents.Labelize(cols, idx, labels)

	return &data
}

// Customize returns a new model with customized column layout.
func (t *TableData) Customize(vs *config.ViewSetting, sc SortColumn, manual, wide bool) (*TableData, SortColumn) {
	if vs.IsBlank() {
		if sc.Name != "" {
			return t, sc
		}
		psc, err := t.sortCol(vs)
		if err == nil {
			return t, psc
		}
		return t, sc
	}

	cols := vs.Columns
	cdata := TableData{
		gvr:       t.gvr,
		namespace: t.namespace,
		header:    t.header.Customize(cols, wide),
	}
	ids := t.header.MapIndices(cols, wide)
	cdata.rowEvents = t.rowEvents.Customize(ids)
	if manual || vs == nil {
		return &cdata, sc
	}
	psc, err := cdata.sortCol(vs)
	if err != nil {
		return &cdata, sc
	}

	return &cdata, psc
}

func (t *TableData) sortCol(vs *config.ViewSetting) (SortColumn, error) {
	var psc SortColumn

	if t.HeaderCount() == 0 {
		return psc, errors.New("no header found")
	}
	name, order, _ := vs.SortCol()
	if _, ok := t.header.IndexOf(name, false); ok {
		psc.Name, psc.ASC = name, order
		return psc, nil
	}
	if client.IsAllNamespaces(t.GetNamespace()) {
		if _, ok := t.header.IndexOf("NAMESPACE", false); ok {
			psc.Name = "NAMESPACE"
		} else if _, ok := t.header.IndexOf("NAME", false); ok {
			psc.Name = "NAME"
		}
	} else {
		if _, ok := t.header.IndexOf("NAME", false); ok {
			psc.Name = "NAME"
		} else {
			psc.Name = t.header[0].Name
		}
	}
	psc.ASC = true

	return psc, nil
}

// Clear clears out the entire table.
func (t *TableData) Clear() {
	t.mx.Lock()
	defer t.mx.Unlock()

	t.header = t.header.Clear()
	t.rowEvents.Clear()
}

// Clone returns a copy of the table.
func (t *TableData) Clone() *TableData {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return &TableData{
		header:    t.header.Clone(),
		rowEvents: t.rowEvents.Clone(),
		namespace: t.namespace,
		gvr:       t.gvr,
	}
}

func (t *TableData) ColumnNames(w bool) []string {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return t.header.ColumnNames(w)
}

// GetHeader returns table header.
func (t *TableData) GetHeader() Header {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return t.header
}

// SetHeader sets table header.
func (t *TableData) SetHeader(ns string, h Header) {
	t.mx.Lock()
	defer t.mx.Unlock()

	t.namespace, t.header = ns, h
}

// Update computes row deltas and update the table data.
func (t *TableData) Update(rows Rows) {
	empty := t.Empty()
	kk := make(map[string]struct{}, len(rows))
	var blankDelta DeltaRow
	t.mx.Lock()
	{
		for _, row := range rows {
			kk[row.ID] = struct{}{}
			if empty {
				t.rowEvents.Add(NewRowEvent(EventAdd, row))
				continue
			}
			if index, ok := t.rowEvents.FindIndex(row.ID); ok {
				ev, ok := t.rowEvents.At(index)
				if !ok {
					continue
				}
				delta := NewDeltaRow(ev.Row, row, t.header)
				if delta.IsBlank() {
					ev.Kind, ev.Deltas, ev.Row = EventUnchanged, blankDelta, row
					t.rowEvents.Set(index, ev)
				} else {
					t.rowEvents.Set(index, NewRowEventWithDeltas(row, delta))
				}
				continue
			}
			t.rowEvents.Add(NewRowEvent(EventAdd, row))
		}
	}
	t.mx.Unlock()

	if !empty {
		t.Delete(kk)
	}
}

// Delete removes items in cache that are no longer valid.
func (t *TableData) Delete(newKeys map[string]struct{}) {
	t.mx.Lock()
	{
		victims := make([]string, 0, 10)
		t.rowEvents.Range(func(_ int, e RowEvent) bool {
			if _, ok := newKeys[e.Row.ID]; !ok {
				victims = append(victims, e.Row.ID)
			} else {
				delete(newKeys, e.Row.ID)
			}
			return true
		})
		for _, id := range victims {
			if err := t.rowEvents.Delete(id); err != nil {
				log.Error().Err(err).Msgf("table delete failed: %q", id)
			}
		}
	}
	t.mx.Unlock()
}

// Diff checks if two tables are equal.
func (t *TableData) Diff(t2 *TableData) bool {
	if t2 == nil || t.namespace != t2.namespace || t.header.Diff(t2.header) {
		return true
	}
	idx, ok := t.header.IndexOf("AGE", true)
	if !ok {
		idx = -1
	}
	return t.rowEvents.Diff(t2.rowEvents, idx)
}
