// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/stretchr/testify/assert"
)

func TestDeltaLabelize(t *testing.T) {
	uu := map[string]struct {
		o model1.Row
		n model1.Row
		e model1.DeltaRow
	}{
		"same": {
			o: model1.Row{
				Fields: model1.Fields{"a", "b", "blee=fred,doh=zorg"},
			},
			n: model1.Row{
				Fields: model1.Fields{"a", "b", "blee=fred1,doh=zorg"},
			},
			e: model1.DeltaRow{"", "", "fred", "zorg"},
		},
	}

	hh := model1.Header{
		model1.HeaderColumn{Name: "A"},
		model1.HeaderColumn{Name: "B"},
		model1.HeaderColumn{Name: "C"},
	}
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			d := model1.NewDeltaRow(u.o, u.n, hh)
			d = d.Labelize([]int{0, 1}, 2)
			assert.Equal(t, u.e, d)
		})
	}
}

func TestDeltaCustomize(t *testing.T) {
	uu := map[string]struct {
		r1, r2 model1.Row
		cols   []int
		e      model1.DeltaRow
	}{
		"same": {
			r1: model1.Row{
				Fields: model1.Fields{"a", "b", "c"},
			},
			r2: model1.Row{
				Fields: model1.Fields{"a", "b", "c"},
			},
			cols: []int{0, 1, 2},
			e:    model1.DeltaRow{"", "", ""},
		},
		"empty": {
			r1: model1.Row{
				Fields: model1.Fields{"a", "b", "c"},
			},
			r2: model1.Row{
				Fields: model1.Fields{"a", "b", "c"},
			},
			e: model1.DeltaRow{},
		},
		"diff-full": {
			r1: model1.Row{
				Fields: model1.Fields{"a", "b", "c"},
			},
			r2: model1.Row{
				Fields: model1.Fields{"a1", "b1", "c1"},
			},
			cols: []int{0, 1, 2},
			e:    model1.DeltaRow{"a", "b", "c"},
		},
		"diff-reverse": {
			r1: model1.Row{
				Fields: model1.Fields{"a", "b", "c"},
			},
			r2: model1.Row{
				Fields: model1.Fields{"a1", "b1", "c1"},
			},
			cols: []int{2, 1, 0},
			e:    model1.DeltaRow{"c", "b", "a"},
		},
		"diff-skip": {
			r1: model1.Row{
				Fields: model1.Fields{"a", "b", "c"},
			},
			r2: model1.Row{
				Fields: model1.Fields{"a1", "b1", "c1"},
			},
			cols: []int{2, 0},
			e:    model1.DeltaRow{"c", "a"},
		},
		"diff-missing": {
			r1: model1.Row{
				Fields: model1.Fields{"a", "b", "c"},
			},
			r2: model1.Row{
				Fields: model1.Fields{"a1", "b1", "c1"},
			},
			cols: []int{2, 10, 0},
			e:    model1.DeltaRow{"c", "", "a"},
		},
		"diff-negative": {
			r1: model1.Row{
				Fields: model1.Fields{"a", "b", "c"},
			},
			r2: model1.Row{
				Fields: model1.Fields{"a1", "b1", "c1"},
			},
			cols: []int{2, -1, 0},
			e:    model1.DeltaRow{"c", "", "a"},
		},
	}

	hh := model1.Header{
		model1.HeaderColumn{Name: "A"},
		model1.HeaderColumn{Name: "B"},
		model1.HeaderColumn{Name: "C"},
	}
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			d := model1.NewDeltaRow(u.r1, u.r2, hh)
			out := make(model1.DeltaRow, len(u.cols))
			d.Customize(u.cols, out)
			assert.Equal(t, u.e, out)
		})
	}
}

func TestDeltaNew(t *testing.T) {
	uu := map[string]struct {
		o     model1.Row
		n     model1.Row
		blank bool
		e     model1.DeltaRow
	}{
		"same": {
			o: model1.Row{
				Fields: model1.Fields{"a", "b", "c"},
			},
			n: model1.Row{
				Fields: model1.Fields{"a", "b", "c"},
			},
			blank: true,
			e:     model1.DeltaRow{"", "", ""},
		},
		"diff": {
			o: model1.Row{
				Fields: model1.Fields{"a1", "b", "c"},
			},
			n: model1.Row{
				Fields: model1.Fields{"a", "b", "c"},
			},
			e: model1.DeltaRow{"a1", "", ""},
		},
		"diff2": {
			o: model1.Row{
				Fields: model1.Fields{"a", "b", "c"},
			},
			n: model1.Row{
				Fields: model1.Fields{"a", "b1", "c"},
			},
			e: model1.DeltaRow{"", "b", ""},
		},
		"diffLast": {
			o: model1.Row{
				Fields: model1.Fields{"a", "b", "c"},
			},
			n: model1.Row{
				Fields: model1.Fields{"a", "b", "c1"},
			},
			e: model1.DeltaRow{"", "", "c"},
		},
	}

	hh := model1.Header{
		model1.HeaderColumn{Name: "A"},
		model1.HeaderColumn{Name: "B"},
		model1.HeaderColumn{Name: "C"},
	}
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			d := model1.NewDeltaRow(u.o, u.n, hh)
			assert.Equal(t, u.e, d)
			assert.Equal(t, u.blank, d.IsBlank())
		})
	}
}

func TestDeltaBlank(t *testing.T) {
	uu := map[string]struct {
		r model1.DeltaRow
		e bool
	}{
		"empty": {
			r: model1.DeltaRow{},
			e: true,
		},
		"blank": {
			r: model1.DeltaRow{"", "", ""},
			e: true,
		},
		"notblank": {
			r: model1.DeltaRow{"", "", "z"},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.r.IsBlank())
		})
	}
}

func TestDeltaDiff(t *testing.T) {
	uu := map[string]struct {
		d1, d2 model1.DeltaRow
		ageCol int
		e      bool
	}{
		"empty": {
			d1:     model1.DeltaRow{"f1", "f2", "f3"},
			ageCol: 2,
			e:      true,
		},
		"same": {
			d1:     model1.DeltaRow{"f1", "f2", "f3"},
			d2:     model1.DeltaRow{"f1", "f2", "f3"},
			ageCol: -1,
		},
		"diff": {
			d1:     model1.DeltaRow{"f1", "f2", "f3"},
			d2:     model1.DeltaRow{"f1", "f2", "f13"},
			ageCol: -1,
			e:      true,
		},
		"diff-age-first": {
			d1:     model1.DeltaRow{"f1", "f2", "f3"},
			d2:     model1.DeltaRow{"f1", "f2", "f13"},
			ageCol: 0,
			e:      true,
		},
		"diff-age-last": {
			d1:     model1.DeltaRow{"f1", "f2", "f3"},
			d2:     model1.DeltaRow{"f1", "f2", "f13"},
			ageCol: 2,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.d1.Diff(u.d2, u.ageCol))
		})
	}
}
