package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
)

func TestAliasColorer(t *testing.T) {
	var a render.Alias

	r := render.Row{ID: "g/v/r", Fields: render.Fields{"r", "blee", "g"}}
	uu := map[string]struct {
		ns string
		re render.RowEvent
		e  tcell.Color
	}{
		"addAll": {
			ns: render.AllNamespaces,
			re: render.RowEvent{Kind: render.EventAdd, Row: r},
			e:  tcell.ColorMediumSpringGreen},
		"deleteAll": {
			ns: render.AllNamespaces,
			re: render.RowEvent{Kind: render.EventDelete, Row: r},
			e:  tcell.ColorMediumSpringGreen},
		"updateAll": {
			ns: render.AllNamespaces,
			re: render.RowEvent{Kind: render.EventUpdate, Row: r},
			e:  tcell.ColorMediumSpringGreen,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, a.ColorerFunc()(u.ns, u.re))
		})
	}
}

func TestAliasHeader(t *testing.T) {
	h := render.HeaderRow{
		render.Header{Name: "RESOURCE"},
		render.Header{Name: "COMMAND"},
		render.Header{Name: "APIGROUP"},
	}

	var a render.Alias
	assert.Equal(t, h, a.Header("fred"))
	assert.Equal(t, h, a.Header(render.AllNamespaces))
}

func TestAliasRender(t *testing.T) {
	a := render.Alias{}

	o := render.AliasRes{
		GVR:     "fred/v1/blee",
		Aliases: []string{"a", "b", "c"},
	}

	var r render.Row
	assert.Nil(t, a.Render(o, "fred/v1/blee", &r))
	assert.Equal(t, render.Row{ID: "fred/v1/blee", Fields: render.Fields{"blee", "a,b,c", "fred"}}, r)
}

func BenchmarkAlias(b *testing.B) {
	o := render.AliasRes{
		GVR:     "fred/v1/blee",
		Aliases: []string{"a", "b", "c"},
	}
	var a render.Alias

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		var r render.Row
		a.Render(o, "aliases", &r)
	}
}
