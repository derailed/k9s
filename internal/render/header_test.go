package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestHeaderMapIndices(t *testing.T) {
	uu := map[string]struct {
		h1   render.Header
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
		h    render.Header
		name string
		wide bool
		e    int
	}{
		"shown": {
			h:    makeHeader(),
			name: "A",
			e:    0,
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
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.h.IndexOf(u.name, u.wide))
		})
	}
}

func TestHeaderCustomize(t *testing.T) {
	uu := map[string]struct {
		h    render.Header
		cols []string
		wide bool
		e    render.Header
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
			h: render.Header{
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "B", Wide: true},
				render.HeaderColumn{Name: "C"},
			},
			cols: []string{"C", "A"},
			e: render.Header{
				render.HeaderColumn{Name: "C"},
				render.HeaderColumn{Name: "A"},
			},
		},
		"reverse-wide": {
			h: render.Header{
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "B", Wide: true},
				render.HeaderColumn{Name: "C"},
			},
			cols: []string{"C", "A"},
			wide: true,
			e: render.Header{
				render.HeaderColumn{Name: "C"},
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "B", Wide: true},
			},
		},
		"toggle-wide": {
			h: render.Header{
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "B", Wide: true},
				render.HeaderColumn{Name: "C"},
			},
			cols: []string{"C", "B"},
			wide: true,
			e: render.Header{
				render.HeaderColumn{Name: "C"},
				render.HeaderColumn{Name: "B", Wide: false},
				render.HeaderColumn{Name: "A", Wide: true},
			},
		},
		"missing": {
			h: render.Header{
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "B", Wide: true},
				render.HeaderColumn{Name: "C"},
			},
			cols: []string{"BLEE", "A"},
			wide: true,
			e: render.Header{
				render.HeaderColumn{Name: "BLEE"},
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "B", Wide: true},
				render.HeaderColumn{Name: "C", Wide: true},
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
		h1, h2 render.Header
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
			h1: render.Header{
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "B", Wide: true},
				render.HeaderColumn{Name: "C"},
			},
			h2: render.Header{
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "B"},
				render.HeaderColumn{Name: "C"},
			},
			e: true,
		},
		"differ-order": {
			h1: render.Header{
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "B", Wide: true},
				render.HeaderColumn{Name: "C"},
			},
			h2: render.Header{
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "C"},
				render.HeaderColumn{Name: "B", Wide: true},
			},
			e: true,
		},
		"differ-name": {
			h1: render.Header{
				render.HeaderColumn{Name: "A"},
			},
			h2: render.Header{
				render.HeaderColumn{Name: "B"},
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
		h      render.Header
		age, e bool
	}{
		"no-age": {
			h: render.Header{},
		},
		"age": {
			h: render.Header{
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "B", Wide: true},
				render.HeaderColumn{Name: "AGE", Time: true},
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

func TestHeaderValidColIndex(t *testing.T) {
	uu := map[string]struct {
		h render.Header
		e int
	}{
		"none": {
			h: render.Header{},
			e: -1,
		},
		"valid": {
			h: render.Header{
				render.HeaderColumn{Name: "A"},
				render.HeaderColumn{Name: "B", Wide: true},
				render.HeaderColumn{Name: "VALID", Wide: true},
			},
			e: 2,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.h.ValidColIndex())
		})
	}
}

func TestHeaderColumns(t *testing.T) {
	uu := map[string]struct {
		h    render.Header
		wide bool
		e    []string
	}{
		"empty": {
			h: render.Header{},
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
			assert.Equal(t, u.e, u.h.Columns(u.wide))
		})
	}
}

func TestHeaderClone(t *testing.T) {
	uu := map[string]struct {
		h render.Header
	}{
		"empty": {
			h: render.Header{},
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

func makeHeader() render.Header {
	return render.Header{
		render.HeaderColumn{Name: "A"},
		render.HeaderColumn{Name: "B", Wide: true},
		render.HeaderColumn{Name: "C"},
	}
}
