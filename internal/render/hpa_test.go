// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"testing"

	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestHorizontalPodAutoscalerColorer(t *testing.T) {
	hpaHeader := Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "REFERENCE"},
		HeaderColumn{Name: "TARGETS%"},
		HeaderColumn{Name: "MINPODS", Align: tview.AlignRight},
		HeaderColumn{Name: "MAXPODS", Align: tview.AlignRight},
		HeaderColumn{Name: "REPLICAS", Align: tview.AlignRight},
		HeaderColumn{Name: "AGE", Time: true},
	}

	uu := map[string]struct {
		h  Header
		re RowEvent
		e  tcell.Color
	}{
		"when replicas = maxpods": {
			h: hpaHeader,
			re: RowEvent{
				Kind: EventUnchanged,
				Row: Row{
					Fields: Fields{"blee", "fred", "fred", "100%", "1", "5", "5", "1d"},
				},
			},
			e: ErrColor,
		},
		"when replicas > maxpods, for some reason": {
			h: hpaHeader,
			re: RowEvent{
				Kind: EventUnchanged,
				Row: Row{
					Fields: Fields{"blee", "fred", "fred", "100%", "1", "5", "6", "1d"},
				},
			},
			e: ErrColor,
		},
		"when replicas < maxpods": {
			h: hpaHeader,
			re: RowEvent{
				Kind: EventUnchanged,
				Row: Row{
					Fields: Fields{"blee", "fred", "fred", "100%", "1", "5", "1", "1d"},
				},
			},
			e: StdColor,
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
