// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestHorizontalPodAutoscalerColorer(t *testing.T) {
	hpaHeader := model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "REFERENCE"},
		model1.HeaderColumn{Name: "TARGETS%"},
		model1.HeaderColumn{Name: "MINPODS", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "MAXPODS", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "REPLICAS", Align: tview.AlignRight},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}

	uu := map[string]struct {
		h  model1.Header
		re *model1.RowEvent
		e  tcell.Color
	}{
		"when replicas = maxpods": {
			h: hpaHeader,
			re: &model1.RowEvent{
				Kind: model1.EventUnchanged,
				Row: model1.Row{
					Fields: model1.Fields{"blee", "fred", "fred", "100%", "1", "5", "5", "1d"},
				},
			},
			e: model1.ErrColor,
		},
		"when replicas > maxpods, for some reason": {
			h: hpaHeader,
			re: &model1.RowEvent{
				Kind: model1.EventUnchanged,
				Row: model1.Row{
					Fields: model1.Fields{"blee", "fred", "fred", "100%", "1", "5", "6", "1d"},
				},
			},
			e: model1.ErrColor,
		},
		"when replicas < maxpods": {
			h: hpaHeader,
			re: &model1.RowEvent{
				Kind: model1.EventUnchanged,
				Row: model1.Row{
					Fields: model1.Fields{"blee", "fred", "fred", "100%", "1", "5", "1", "1d"},
				},
			},
			e: model1.StdColor,
		},
	}

	var r HorizontalPodAutoscaler
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, r.ColorerFunc()("", u.h, u.re))
		})
	}
}
