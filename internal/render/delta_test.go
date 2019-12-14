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

	for k := range uu {
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			d := render.NewDeltaRow(uc.o, uc.n, false)
			assert.Equal(t, uc.e, d)
			assert.Equal(t, uc.blank, d.IsBlank())
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
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, uc.e, uc.r.IsBlank())
		})
	}
}
