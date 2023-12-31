// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestHint(t *testing.T) {
	uu := map[string]struct {
		hh model.MenuHints
		e  int
	}{
		"none": {
			model.MenuHints{},
			0,
		},
		"hints": {
			model.MenuHints{
				{Mnemonic: "a", Description: "blee"},
				{Mnemonic: "b", Description: "fred"},
			},
			2,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			h := model.NewHint()
			l := hintL{count: -1}
			h.AddListener(&l)
			h.SetHints(u.hh)

			assert.Equal(t, u.e, l.count)
			assert.Equal(t, u.e, len(h.Peek()))
		})
	}
}

func TestHintRemoveListener(t *testing.T) {
	h := model.NewHint()
	l1, l2, l3 := &hintL{}, &hintL{}, &hintL{}
	h.AddListener(l1)
	h.AddListener(l2)

	h.RemoveListener(l2)
	h.RemoveListener(l3)
	h.RemoveListener(l1)

	h.SetHints(model.MenuHints{
		model.MenuHint{Mnemonic: "a", Description: "Blee"},
	})

	assert.Equal(t, 0, l1.count)
	assert.Equal(t, 0, l2.count)
	assert.Equal(t, 0, l3.count)
}

// ----------------------------------------------------------------------------
// Helpers...

type hintL struct {
	count int
}

func (h *hintL) HintsChanged(hh model.MenuHints) {
	h.count++
	h.count += len(hh)
}
