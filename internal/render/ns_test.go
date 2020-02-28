package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
)

func TestNSColorer(t *testing.T) {
	uu := map[string]struct {
		re render.RowEvent
		e  tcell.Color
	}{
		"add": {
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{
						"blee",
						"Active",
					},
				},
			},
			e: render.AddColor,
		},
		"update": {
			re: render.RowEvent{
				Kind: render.EventUpdate,
				Row: render.Row{
					Fields: render.Fields{
						"blee",
						"Active",
					},
				},
			},
			e: render.StdColor,
		},
		"decorator": {
			re: render.RowEvent{
				Kind: render.EventAdd,
				Row: render.Row{
					Fields: render.Fields{
						"blee*",
						"Active",
					},
				},
			},
			e: render.HighlightColor,
		},
	}

	h := render.Header{
		render.HeaderColumn{Name: "NAME"},
		render.HeaderColumn{Name: "STATUS"},
	}

	var r render.Namespace
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, r.ColorerFunc()("", h, u.re))
		})
	}
}

func TestNamespaceRender(t *testing.T) {
	c := render.Namespace{}
	r := render.NewRow(3)
	c.Render(load(t, "ns"), "-", &r)

	assert.Equal(t, "-/kube-system", r.ID)
	assert.Equal(t, render.Fields{"kube-system", "Active"}, r.Fields[:2])
}
