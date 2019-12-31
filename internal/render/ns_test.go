package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestNSColorer(t *testing.T) {
	var (
		ns   = render.Row{Fields: render.Fields{"blee", "Active"}}
		term = render.Row{Fields: render.Fields{"blee", render.Terminating}}
		dead = render.Row{Fields: render.Fields{"blee", "Inactive"}}
	)

	uu := colorerUCs{
		// Add AllNS
		{"", render.RowEvent{Kind: render.EventAdd, Row: ns}, render.AddColor},
		// Mod AllNS
		{"", render.RowEvent{Kind: render.EventUpdate, Row: ns}, render.ModColor},
		// MoChange AllNS
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: ns}, render.StdColor},
		// Bust NS
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: term}, render.ErrColor},
		// Bust NS
		{"", render.RowEvent{Kind: render.EventUnchanged, Row: dead}, render.ErrColor},
	}

	var n render.Namespace
	f := n.ColorerFunc()
	for _, u := range uu {
		assert.Equal(t, u.e, f(u.ns, u.r))
	}
}

func TestNamespaceRender(t *testing.T) {
	c := render.Namespace{}
	r := render.NewRow(3)
	c.Render(load(t, "ns"), "-", &r)

	assert.Equal(t, "kube-system", r.ID)
	assert.Equal(t, render.Fields{"kube-system", "Active"}, r.Fields[:2])
}
