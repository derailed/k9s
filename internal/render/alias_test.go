// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestAliasColorer(t *testing.T) {
	var a render.Alias
	h := model1.Header{
		model1.HeaderColumn{Name: "A"},
		model1.HeaderColumn{Name: "B"},
		model1.HeaderColumn{Name: "C"},
	}
	r := model1.Row{ID: "g/v/r", Fields: model1.Fields{"r", "blee", "g"}}
	uu := map[string]struct {
		ns string
		re model1.RowEvent
		e  tcell.Color
	}{
		"addAll": {
			ns: client.NamespaceAll,
			re: model1.RowEvent{Kind: model1.EventAdd, Row: r},
			e:  tcell.ColorBlue,
		},
		"deleteAll": {
			ns: client.NamespaceAll,
			re: model1.RowEvent{Kind: model1.EventDelete, Row: r},
			e:  tcell.ColorGray,
		},
		"updateAll": {
			ns: client.NamespaceAll,
			re: model1.RowEvent{Kind: model1.EventUpdate, Row: r},
			e:  tcell.ColorDefault,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, a.ColorerFunc()(u.ns, h, &u.re))
		})
	}
}

func TestAliasHeader(t *testing.T) {
	h := model1.Header{
		model1.HeaderColumn{Name: "RESOURCE"},
		model1.HeaderColumn{Name: "COMMAND"},
		model1.HeaderColumn{Name: "API-GROUP"},
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

	var r model1.Row
	assert.Nil(t, a.Render(o, "fred/v1/blee", &r))
	assert.Equal(t, model1.Row{ID: "fred/v1/blee", Fields: model1.Fields{"blee", "a,b,c", "fred"}}, r)
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
		var r model1.Row
		_ = a.Render(o, "aliases", &r)
	}
}
