// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestDefaultColorer(t *testing.T) {
	uu := map[string]struct {
		re model1.RowEvent
		e  tcell.Color
	}{
		"add": {
			model1.RowEvent{
				Kind: model1.EventAdd,
			},
			model1.AddColor,
		},
		"update": {
			model1.RowEvent{
				Kind: model1.EventUpdate,
			},
			model1.ModColor,
		},
		"delete": {
			model1.RowEvent{
				Kind: model1.EventDelete,
			},
			model1.KillColor,
		},
		"no-change": {
			model1.RowEvent{
				Kind: model1.EventUnchanged,
			},
			model1.StdColor,
		},
		"invalid": {
			model1.RowEvent{
				Kind: model1.EventUnchanged,
				Row: model1.Row{
					Fields: model1.Fields{"", "", "blah"},
				},
			},
			model1.ErrColor,
		},
	}

	h := model1.Header{
		model1.HeaderColumn{Name: "A"},
		model1.HeaderColumn{Name: "B"},
		model1.HeaderColumn{Name: "VALID"},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, model1.DefaultColorer("", h, &u.re))
		})
	}
}
