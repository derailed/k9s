package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestDefaultColorer(t *testing.T) {
	uu := map[string]struct {
		re render.RowEvent
		e  tcell.Color
	}{
		"add": {
			render.RowEvent{
				Kind: render.EventAdd,
			},
			render.AddColor,
		},
		"update": {
			render.RowEvent{
				Kind: render.EventUpdate,
			},
			render.ModColor,
		},
		"delete": {
			render.RowEvent{
				Kind: render.EventDelete,
			},
			render.KillColor,
		},
		"no-change": {
			render.RowEvent{
				Kind: render.EventUnchanged,
			},
			render.StdColor,
		},
		"invalid": {
			render.RowEvent{
				Kind: render.EventUnchanged,
				Row: render.Row{
					Fields: render.Fields{"", "", "blah"},
				},
			},
			render.ErrColor,
		},
	}

	h := render.Header{
		render.HeaderColumn{Name: "A"},
		render.HeaderColumn{Name: "B"},
		render.HeaderColumn{Name: "VALID"},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, render.DefaultColorer("", h, u.re))
		})
	}
}
