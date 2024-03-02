// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/stretchr/testify/assert"
)

func TestHeaderMapIndices(t *testing.T) {
	uu := map[string]struct {
		h1   model1.Header
		cols []string
		wide bool
		e    []int
	}{
		"all": {
			h1:   makeHeader(),
			cols: []string{"A", "B", "C"},
			e:    []int{0, 1, 2},
		},
		"reverse": {
			h1:   makeHeader(),
			cols: []string{"C", "B", "A"},
			e:    []int{2, 1, 0},
		},
		"missing": {
			h1:   makeHeader(),
			cols: []string{"Duh", "B", "A"},
			e:    []int{-1, 1, 0},
		},
		"skip": {
			h1:   makeHeader(),
			cols: []string{"C", "A"},
			e:    []int{2, 0},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			ii := u.h1.MapIndices(u.cols, u.wide)
			assert.Equal(t, u.e, ii)
		})
	}
}

func TestHeaderIndexOf(t *testing.T) {
	uu := map[string]struct {
		h        model1.Header
		name     string
		wide, ok bool
		e        int
	}{
		"shown": {
			h:    makeHeader(),
			name: "A",
			e:    0,
			ok:   true,
		},
		"hidden": {
			h:    makeHeader(),
			name: "B",
			e:    -1,
		},
		"hidden-wide": {
			h:    makeHeader(),
			name: "B",
			wide: true,
			e:    1,
			ok:   true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			idx, ok := u.h.IndexOf(u.name, u.wide)
			assert.Equal(t, u.ok, ok)
			assert.Equal(t, u.e, idx)
		})
	}
}

func TestHeaderCustomize(t *testing.T) {
	uu := map[string]struct {
		h    model1.Header
		cols []string
		wide bool
		e    model1.Header
	}{
		"default": {
			h: makeHeader(),
			e: makeHeader(),
		},
		"default-wide": {
			h:    makeHeader(),
			wide: true,
			e:    makeHeader(),
		},
		"reverse": {
			h: model1.Header{
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "B", Wide: true},
				model1.HeaderColumn{Name: "C"},
			},
			cols: []string{"C", "A"},
			e: model1.Header{
				model1.HeaderColumn{Name: "C"},
				model1.HeaderColumn{Name: "A"},
			},
		},
		"reverse-wide": {
			h: model1.Header{
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "B", Wide: true},
				model1.HeaderColumn{Name: "C"},
			},
			cols: []string{"C", "A"},
			wide: true,
			e: model1.Header{
				model1.HeaderColumn{Name: "C"},
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "B", Wide: true},
			},
		},
		"toggle-wide": {
			h: model1.Header{
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "B", Wide: true},
				model1.HeaderColumn{Name: "C"},
			},
			cols: []string{"C", "B"},
			wide: true,
			e: model1.Header{
				model1.HeaderColumn{Name: "C"},
				model1.HeaderColumn{Name: "B", Wide: false},
				model1.HeaderColumn{Name: "A", Wide: true},
			},
		},
		"missing": {
			h: model1.Header{
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "B", Wide: true},
				model1.HeaderColumn{Name: "C"},
			},
			cols: []string{"BLEE", "A"},
			wide: true,
			e: model1.Header{
				model1.HeaderColumn{Name: "BLEE"},
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "B", Wide: true},
				model1.HeaderColumn{Name: "C", Wide: true},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.h.Customize(u.cols, u.wide))
		})
	}
}

func TestHeaderDiff(t *testing.T) {
	uu := map[string]struct {
		h1, h2 model1.Header
		e      bool
	}{
		"same": {
			h1: makeHeader(),
			h2: makeHeader(),
		},
		"size": {
			h1: makeHeader(),
			h2: makeHeader()[1:],
			e:  true,
		},
		"differ-wide": {
			h1: model1.Header{
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "B", Wide: true},
				model1.HeaderColumn{Name: "C"},
			},
			h2: model1.Header{
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "B"},
				model1.HeaderColumn{Name: "C"},
			},
			e: true,
		},
		"differ-order": {
			h1: model1.Header{
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "B", Wide: true},
				model1.HeaderColumn{Name: "C"},
			},
			h2: model1.Header{
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "C"},
				model1.HeaderColumn{Name: "B", Wide: true},
			},
			e: true,
		},
		"differ-name": {
			h1: model1.Header{
				model1.HeaderColumn{Name: "A"},
			},
			h2: model1.Header{
				model1.HeaderColumn{Name: "B"},
			},
			e: true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.h1.Diff(u.h2))
		})
	}
}

func TestHeaderHasAge(t *testing.T) {
	uu := map[string]struct {
		h      model1.Header
		age, e bool
	}{
		"no-age": {
			h: model1.Header{},
		},
		"age": {
			h: model1.Header{
				model1.HeaderColumn{Name: "A"},
				model1.HeaderColumn{Name: "B", Wide: true},
				model1.HeaderColumn{Name: "AGE", Time: true},
			},
			e:   true,
			age: true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.h.HasAge())
			assert.Equal(t, u.e, u.h.IsTimeCol(2))
		})
	}
}

func TestHeaderColumns(t *testing.T) {
	uu := map[string]struct {
		h    model1.Header
		wide bool
		e    []string
	}{
		"empty": {
			h: model1.Header{},
		},
		"regular": {
			h: makeHeader(),
			e: []string{"A", "C"},
		},
		"wide": {
			h:    makeHeader(),
			e:    []string{"A", "B", "C"},
			wide: true,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.h.ColumnNames(u.wide))
		})
	}
}

func TestHeaderClone(t *testing.T) {
	uu := map[string]struct {
		h model1.Header
	}{
		"empty": {
			h: model1.Header{},
		},
		"full": {
			h: makeHeader(),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := u.h.Clone()
			assert.Equal(t, len(u.h), len(c))
			if len(u.h) > 0 {
				u.h[0].Name = "blee"
				assert.Equal(t, "A", c[0].Name)
			}
		})
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeHeader() model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "A"},
		model1.HeaderColumn{Name: "B", Wide: true},
		model1.HeaderColumn{Name: "C"},
	}
}
