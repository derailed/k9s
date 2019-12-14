package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestDelta(t *testing.T) {
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

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			d := render.NewDeltaRow(u.o, u.n, false)
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

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.r.IsBlank())
		})
	}
}
