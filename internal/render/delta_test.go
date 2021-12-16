package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestDeltaLabelize(t *testing.T) {
	uu := map[string]struct {
		o render.Row
		n render.Row
		e render.DeltaRow
	}{
		"same": {
			o: render.Row{
				Fields: render.Fields{"a", "b", "blee=fred,doh=zorg"},
			},
			n: render.Row{
				Fields: render.Fields{"a", "b", "blee=fred1,doh=zorg"},
			},
			e: render.DeltaRow{"", "", "fred", "zorg"},
		},
	}

	hh := render.Header{
		render.HeaderColumn{Name: "A"},
		render.HeaderColumn{Name: "B"},
		render.HeaderColumn{Name: "C"},
	}
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			d := render.NewDeltaRow(u.o, u.n, hh)
			d = d.Labelize([]int{0, 1}, 2)
			assert.Equal(t, u.e, d)
		})
	}
}

func TestDeltaCustomize(t *testing.T) {
	uu := map[string]struct {
		r1, r2 render.Row
		cols   []int
		e      render.DeltaRow
	}{
		"same": {
			r1: render.Row{
				Fields: render.Fields{"a", "b", "c"},
			},
			r2: render.Row{
				Fields: render.Fields{"a", "b", "c"},
			},
			cols: []int{0, 1, 2},
			e:    render.DeltaRow{"", "", ""},
		},
		"empty": {
			r1: render.Row{
				Fields: render.Fields{"a", "b", "c"},
			},
			r2: render.Row{
				Fields: render.Fields{"a", "b", "c"},
			},
			e: render.DeltaRow{},
		},
		"diff-full": {
			r1: render.Row{
				Fields: render.Fields{"a", "b", "c"},
			},
			r2: render.Row{
				Fields: render.Fields{"a1", "b1", "c1"},
			},
			cols: []int{0, 1, 2},
			e:    render.DeltaRow{"a", "b", "c"},
		},
		"diff-reverse": {
			r1: render.Row{
				Fields: render.Fields{"a", "b", "c"},
			},
			r2: render.Row{
				Fields: render.Fields{"a1", "b1", "c1"},
			},
			cols: []int{2, 1, 0},
			e:    render.DeltaRow{"c", "b", "a"},
		},
		"diff-skip": {
			r1: render.Row{
				Fields: render.Fields{"a", "b", "c"},
			},
			r2: render.Row{
				Fields: render.Fields{"a1", "b1", "c1"},
			},
			cols: []int{2, 0},
			e:    render.DeltaRow{"c", "a"},
		},
		"diff-missing": {
			r1: render.Row{
				Fields: render.Fields{"a", "b", "c"},
			},
			r2: render.Row{
				Fields: render.Fields{"a1", "b1", "c1"},
			},
			cols: []int{2, 10, 0},
			e:    render.DeltaRow{"c", "", "a"},
		},
		"diff-negative": {
			r1: render.Row{
				Fields: render.Fields{"a", "b", "c"},
			},
			r2: render.Row{
				Fields: render.Fields{"a1", "b1", "c1"},
			},
			cols: []int{2, -1, 0},
			e:    render.DeltaRow{"c", "", "a"},
		},
	}

	hh := render.Header{
		render.HeaderColumn{Name: "A"},
		render.HeaderColumn{Name: "B"},
		render.HeaderColumn{Name: "C"},
	}
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			d := render.NewDeltaRow(u.r1, u.r2, hh)
			out := make(render.DeltaRow, len(u.cols))
			d.Customize(u.cols, out)
			assert.Equal(t, u.e, out)
		})
	}
}

func TestDeltaNew(t *testing.T) {
	uu := map[string]struct {
		o     render.Row
		n     render.Row
		blank bool
		e     render.DeltaRow
	}{
		"same": {
			o: render.Row{
				Fields: render.Fields{"a", "b", "c"},
			},
			n: render.Row{
				Fields: render.Fields{"a", "b", "c"},
			},
			blank: true,
			e:     render.DeltaRow{"", "", ""},
		},
		"diff": {
			o: render.Row{
				Fields: render.Fields{"a1", "b", "c"},
			},
			n: render.Row{
				Fields: render.Fields{"a", "b", "c"},
			},
			e: render.DeltaRow{"a1", "", ""},
		},
		"diff2": {
			o: render.Row{
				Fields: render.Fields{"a", "b", "c"},
			},
			n: render.Row{
				Fields: render.Fields{"a", "b1", "c"},
			},
			e: render.DeltaRow{"", "b", ""},
		},
		"diffLast": {
			o: render.Row{
				Fields: render.Fields{"a", "b", "c"},
			},
			n: render.Row{
				Fields: render.Fields{"a", "b", "c1"},
			},
			e: render.DeltaRow{"", "", "c"},
		},
	}

	hh := render.Header{
		render.HeaderColumn{Name: "A"},
		render.HeaderColumn{Name: "B"},
		render.HeaderColumn{Name: "C"},
	}
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			d := render.NewDeltaRow(u.o, u.n, hh)
			assert.Equal(t, u.e, d)
			assert.Equal(t, u.blank, d.IsBlank())
		})
	}
}

func TestDeltaBlank(t *testing.T) {
	uu := map[string]struct {
		r render.DeltaRow
		e bool
	}{
		"empty": {
			r: render.DeltaRow{},
			e: true,
		},
		"blank": {
			r: render.DeltaRow{"", "", ""},
			e: true,
		},
		"notblank": {
			r: render.DeltaRow{"", "", "z"},
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
		d1, d2 render.DeltaRow
		ageCol int
		e      bool
	}{
		"empty": {
			d1:     render.DeltaRow{"f1", "f2", "f3"},
			ageCol: 2,
			e:      true,
		},
		"same": {
			d1:     render.DeltaRow{"f1", "f2", "f3"},
			d2:     render.DeltaRow{"f1", "f2", "f3"},
			ageCol: -1,
		},
		"diff": {
			d1:     render.DeltaRow{"f1", "f2", "f3"},
			d2:     render.DeltaRow{"f1", "f2", "f13"},
			ageCol: -1,
			e:      true,
		},
		"diff-age-first": {
			d1:     render.DeltaRow{"f1", "f2", "f3"},
			d2:     render.DeltaRow{"f1", "f2", "f13"},
			ageCol: 0,
			e:      true,
		},
		"diff-age-last": {
			d1:     render.DeltaRow{"f1", "f2", "f3"},
			d2:     render.DeltaRow{"f1", "f2", "f13"},
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
