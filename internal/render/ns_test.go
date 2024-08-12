// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestNSColorer(t *testing.T) {
	uu := map[string]struct {
		re model1.RowEvent
		e  tcell.Color
	}{
		"add": {
			re: model1.RowEvent{
				Kind: model1.EventAdd,
				Row: model1.Row{
					Fields: model1.Fields{
						"blee",
						"Active",
					},
				},
			},
			e: model1.AddColor,
		},
		"update": {
			re: model1.RowEvent{
				Kind: model1.EventUpdate,
				Row: model1.Row{
					Fields: model1.Fields{
						"blee",
						"Active",
					},
				},
			},
			e: model1.StdColor,
		},
		"decorator": {
			re: model1.RowEvent{
				Kind: model1.EventAdd,
				Row: model1.Row{
					Fields: model1.Fields{
						"blee*",
						"Active",
					},
				},
			},
			e: model1.HighlightColor,
		},
	}

	h := model1.Header{
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "STATUS"},
	}

	var r render.Namespace
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, r.ColorerFunc()("", h, &u.re))
		})
	}
}

func TestNamespaceRender(t *testing.T) {
	c := render.Namespace{}
	r := model1.NewRow(3)

	assert.NoError(t, c.Render(load(t, "ns"), "-", &r))
	assert.Equal(t, "-/kube-system", r.ID)
	assert.Equal(t, model1.Fields{"kube-system", "Active"}, r.Fields[:2])
}
