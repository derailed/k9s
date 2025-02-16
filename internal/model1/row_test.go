// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1_test

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/model1"
	"github.com/stretchr/testify/assert"
)

func BenchmarkRowCustomize(b *testing.B) {
	row := model1.Row{ID: "fred", Fields: model1.Fields{"f1", "f2", "f3"}}
	cols := []int{0, 1, 2}
	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = row.Customize(cols)
	}
}

func TestFieldCustomize(t *testing.T) {
	uu := map[string]struct {
		fields model1.Fields
		cols   []int
		e      model1.Fields
	}{
		"empty": {
			fields: model1.Fields{},
			cols:   []int{0, 1, 2},
			e:      model1.Fields{"", "", ""},
		},
		"no-cols": {
			fields: model1.Fields{"f1", "f2", "f3"},
			cols:   []int{},
			e:      model1.Fields{},
		},
		"reverse": {
			fields: model1.Fields{"f1", "f2", "f3"},
			cols:   []int{1, 0},
			e:      model1.Fields{"f2", "f1"},
		},
		"missing": {
			fields: model1.Fields{"f1", "f2", "f3"},
			cols:   []int{10, 0},
			e:      model1.Fields{"", "f1"},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			ff := make(model1.Fields, len(u.cols))
			u.fields.Customize(u.cols, ff)
			assert.Equal(t, u.e, ff)
		})
	}
}

func TestFieldClone(t *testing.T) {
	f := model1.Fields{"a", "b", "c"}
	f1 := f.Clone()

	assert.True(t, reflect.DeepEqual(f, f1))
	assert.NotEqual(t, fmt.Sprintf("%p", f), fmt.Sprintf("%p", f1))
}

func TestRowLabelize(t *testing.T) {
	uu := map[string]struct {
		row  model1.Row
		cols []int
		e    model1.Row
	}{
		"empty": {
			row:  model1.Row{},
			cols: []int{0, 1, 2},
			e:    model1.Row{ID: "", Fields: model1.Fields{"", "", ""}},
		},
		"no-cols-no-data": {
			row:  model1.Row{},
			cols: []int{},
			e:    model1.Row{ID: "", Fields: model1.Fields{}},
		},
		"no-cols-data": {
			row:  model1.Row{ID: "fred", Fields: model1.Fields{"f1", "f2", "f3"}},
			cols: []int{},
			e:    model1.Row{ID: "fred", Fields: model1.Fields{}},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			row := u.row.Customize(u.cols)
			assert.Equal(t, u.e, row)
		})
	}
}

func TestRowCustomize(t *testing.T) {
	uu := map[string]struct {
		row  model1.Row
		cols []int
		e    model1.Row
	}{
		"empty": {
			row:  model1.Row{},
			cols: []int{0, 1, 2},
			e:    model1.Row{ID: "", Fields: model1.Fields{"", "", ""}},
		},
		"no-cols-no-data": {
			row:  model1.Row{},
			cols: []int{},
			e:    model1.Row{ID: "", Fields: model1.Fields{}},
		},
		"no-cols-data": {
			row:  model1.Row{ID: "fred", Fields: model1.Fields{"f1", "f2", "f3"}},
			cols: []int{},
			e:    model1.Row{ID: "fred", Fields: model1.Fields{}},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			row := u.row.Customize(u.cols)
			assert.Equal(t, u.e, row)
		})
	}
}

func TestRowsDelete(t *testing.T) {
	uu := map[string]struct {
		rows model1.Rows
		id   string
		e    model1.Rows
	}{
		"first": {
			rows: model1.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
				{ID: "b", Fields: []string{"albert", "blee"}},
			},
			id: "a",
			e: model1.Rows{
				{ID: "b", Fields: []string{"albert", "blee"}},
			},
		},
		"last": {
			rows: model1.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
				{ID: "b", Fields: []string{"albert", "blee"}},
			},
			id: "b",
			e: model1.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
			},
		},
		"middle": {
			rows: model1.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
				{ID: "b", Fields: []string{"albert", "blee"}},
				{ID: "c", Fields: []string{"fred", "zorg"}},
			},
			id: "b",
			e: model1.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
				{ID: "c", Fields: []string{"fred", "zorg"}},
			},
		},
		"missing": {
			rows: model1.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
				{ID: "b", Fields: []string{"albert", "blee"}},
			},
			id: "zorg",
			e: model1.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
				{ID: "b", Fields: []string{"albert", "blee"}},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			rows := u.rows.Delete(u.id)
			assert.Equal(t, u.e, rows)
		})
	}
}

func TestRowsUpsert(t *testing.T) {
	uu := map[string]struct {
		rows model1.Rows
		row  model1.Row
		e    model1.Rows
	}{
		"add": {
			rows: model1.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
				{ID: "b", Fields: []string{"albert", "blee"}},
			},
			row: model1.Row{ID: "c", Fields: []string{"f1", "f2"}},
			e: model1.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
				{ID: "b", Fields: []string{"albert", "blee"}},
				{ID: "c", Fields: []string{"f1", "f2"}},
			},
		},
		"update": {
			rows: model1.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
				{ID: "b", Fields: []string{"albert", "blee"}},
			},
			row: model1.Row{ID: "a", Fields: []string{"f1", "f2"}},
			e: model1.Rows{
				{ID: "a", Fields: []string{"f1", "f2"}},
				{ID: "b", Fields: []string{"albert", "blee"}},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			rows := u.rows.Upsert(u.row)
			assert.Equal(t, u.e, rows)
		})
	}
}

func TestRowsSortText(t *testing.T) {
	uu := map[string]struct {
		rows     model1.Rows
		col      int
		asc, num bool
		e        model1.Rows
	}{
		"plainAsc": {
			rows: model1.Rows{
				{Fields: []string{"blee", "duh"}},
				{Fields: []string{"albert", "blee"}},
			},
			col: 0,
			asc: true,
			e: model1.Rows{
				{Fields: []string{"albert", "blee"}},
				{Fields: []string{"blee", "duh"}},
			},
		},
		"plainDesc": {
			rows: model1.Rows{
				{Fields: []string{"blee", "duh"}},
				{Fields: []string{"albert", "blee"}},
			},
			col: 0,
			asc: false,
			e: model1.Rows{
				{Fields: []string{"blee", "duh"}},
				{Fields: []string{"albert", "blee"}},
			},
		},
		"numericAsc": {
			rows: model1.Rows{
				{Fields: []string{"10", "duh"}},
				{Fields: []string{"1", "blee"}},
			},
			col: 0,
			num: true,
			asc: true,
			e: model1.Rows{
				{Fields: []string{"1", "blee"}},
				{Fields: []string{"10", "duh"}},
			},
		},
		"numericDesc": {
			rows: model1.Rows{
				{Fields: []string{"10", "duh"}},
				{Fields: []string{"1", "blee"}},
			},
			col: 0,
			num: true,
			asc: false,
			e: model1.Rows{
				{Fields: []string{"10", "duh"}},
				{Fields: []string{"1", "blee"}},
			},
		},
		"composite": {
			rows: model1.Rows{
				{Fields: []string{"blee-duh", "duh"}},
				{Fields: []string{"blee", "blee"}},
			},
			col: 0,
			asc: true,
			e: model1.Rows{
				{Fields: []string{"blee", "blee"}},
				{Fields: []string{"blee-duh", "duh"}},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.rows.Sort(u.col, u.asc, u.num, false, false)
			assert.Equal(t, u.e, u.rows)
		})
	}
}

func TestRowsSortDuration(t *testing.T) {
	uu := map[string]struct {
		rows model1.Rows
		col  int
		asc  bool
		e    model1.Rows
	}{
		"fred": {
			rows: model1.Rows{
				{Fields: []string{"2m24s", "blee"}},
				{Fields: []string{"2m12s", "duh"}},
			},
			col: 0,
			asc: true,
			e: model1.Rows{
				{Fields: []string{"2m12s", "duh"}},
				{Fields: []string{"2m24s", "blee"}},
			},
		},
		"years": {
			rows: model1.Rows{
				{Fields: []string{testTime().Add(-365 * 24 * time.Hour).String(), "blee"}},
				{Fields: []string{testTime().String(), "duh"}},
			},
			col: 0,
			asc: true,
			e: model1.Rows{
				{Fields: []string{testTime().String(), "duh"}},
				{Fields: []string{testTime().Add(-365 * 24 * time.Hour).String(), "blee"}},
			},
		},
		"durationAsc": {
			rows: model1.Rows{
				{Fields: []string{testTime().Add(10 * time.Second).String(), "duh"}},
				{Fields: []string{testTime().String(), "blee"}},
			},
			col: 0,
			asc: true,
			e: model1.Rows{
				{Fields: []string{testTime().String(), "blee"}},
				{Fields: []string{testTime().Add(10 * time.Second).String(), "duh"}},
			},
		},
		"durationDesc": {
			rows: model1.Rows{
				{Fields: []string{testTime().Add(10 * time.Second).String(), "duh"}},
				{Fields: []string{testTime().String(), "blee"}},
			},
			col: 0,
			e: model1.Rows{
				{Fields: []string{testTime().Add(10 * time.Second).String(), "duh"}},
				{Fields: []string{testTime().String(), "blee"}},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.rows.Sort(u.col, u.asc, false, true, false)
			assert.Equal(t, u.e, u.rows)
		})
	}
}

func TestRowsSortMetrics(t *testing.T) {
	uu := map[string]struct {
		rows model1.Rows
		col  int
		asc  bool
		e    model1.Rows
	}{
		"metricAsc": {
			rows: model1.Rows{
				{Fields: []string{"10m", "duh"}},
				{Fields: []string{"1m", "blee"}},
			},
			col: 0,
			asc: true,
			e: model1.Rows{
				{Fields: []string{"1m", "blee"}},
				{Fields: []string{"10m", "duh"}},
			},
		},
		"metricDesc": {
			rows: model1.Rows{
				{Fields: []string{"10000m", "1000Mi"}},
				{Fields: []string{"1m", "50Mi"}},
			},
			col: 1,
			asc: false,
			e: model1.Rows{
				{Fields: []string{"10000m", "1000Mi"}},
				{Fields: []string{"1m", "50Mi"}},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.rows.Sort(u.col, u.asc, true, false, false)
			assert.Equal(t, u.e, u.rows)
		})
	}
}

func TestRowsSortCapacity(t *testing.T) {
	uu := map[string]struct {
		rows model1.Rows
		col  int
		asc  bool
		e    model1.Rows
	}{
		"capacityAsc": {
			rows: model1.Rows{
				{Fields: []string{"10Gi", "duh"}},
				{Fields: []string{"10G", "blee"}},
			},
			col: 0,
			asc: true,
			e: model1.Rows{
				{Fields: []string{"10G", "blee"}},
				{Fields: []string{"10Gi", "duh"}},
			},
		},
		"capacityDesc": {
			rows: model1.Rows{
				{Fields: []string{"10000m", "1000Mi"}},
				{Fields: []string{"1m", "50Mi"}},
			},
			col: 1,
			asc: false,
			e: model1.Rows{
				{Fields: []string{"10000m", "1000Mi"}},
				{Fields: []string{"1m", "50Mi"}},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.rows.Sort(u.col, u.asc, false, false, true)
			assert.Equal(t, u.e, u.rows)
		})
	}
}

func TestLess(t *testing.T) {
	uu := map[string]struct {
		isNumber   bool
		isDuration bool
		isCapacity bool
		id1, id2   string
		v1, v2     string
		e          bool
	}{
		"years": {
			isNumber:   false,
			isDuration: true,
			isCapacity: false,
			id1:        "id1",
			id2:        "id2",
			v1:         "2y263d",
			v2:         "1y179d",
		},
		"hours": {
			isNumber:   false,
			isDuration: true,
			isCapacity: false,
			id1:        "id1",
			id2:        "id2",
			v1:         "2y263d",
			v2:         "19h",
		},
		"capacity1": {
			isNumber:   false,
			isDuration: false,
			isCapacity: true,
			id1:        "id1",
			id2:        "id2",
			v1:         "1Gi",
			v2:         "1G",
			e:          false,
		},
		"capacity2": {
			isNumber:   false,
			isDuration: false,
			isCapacity: true,
			id1:        "id1",
			id2:        "id2",
			v1:         "1Gi",
			v2:         "1Ti",
			e:          true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, model1.Less(u.isNumber, u.isDuration, u.isCapacity, u.id1, u.id2, u.v1, u.v2))
		})
	}
}
