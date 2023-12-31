// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestAliasColorer(t *testing.T) {
	var a render.Alias
	h := render.Header{
		render.HeaderColumn{Name: "A"},
		render.HeaderColumn{Name: "B"},
		render.HeaderColumn{Name: "C"},
	}
	r := render.Row{ID: "g/v/r", Fields: render.Fields{"r", "blee", "g"}}
	uu := map[string]struct {
		ns string
		re render.RowEvent
		e  tcell.Color
	}{
		"addAll": {
			ns: client.NamespaceAll,
			re: render.RowEvent{Kind: render.EventAdd, Row: r},
			e:  tcell.ColorBlue,
		},
		"deleteAll": {
			ns: client.NamespaceAll,
			re: render.RowEvent{Kind: render.EventDelete, Row: r},
			e:  tcell.ColorGray,
		},
		"updateAll": {
			ns: client.NamespaceAll,
			re: render.RowEvent{Kind: render.EventUpdate, Row: r},
			e:  tcell.ColorDefault,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, a.ColorerFunc()(u.ns, h, u.re))
		})
	}
}

func TestAliasHeader(t *testing.T) {
	h := render.Header{
		render.HeaderColumn{Name: "RESOURCE"},
		render.HeaderColumn{Name: "COMMAND"},
		render.HeaderColumn{Name: "API-GROUP"},
	}

	var a render.Alias
	assert.Equal(t, h, a.Header("fred"))
	assert.Equal(t, h, a.Header(client.NamespaceAll))
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
		_ = a.Render(o, "aliases", &r)
	}
}
