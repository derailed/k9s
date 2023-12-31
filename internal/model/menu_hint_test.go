// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model_test

import (
	"sort"
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestMenuHintsSort(t *testing.T) {
	uu := map[string]struct {
		hh model.MenuHints
		e  []int
	}{
		"mixed": {
			hh: model.MenuHints{
				model.MenuHint{Mnemonic: "2", Description: "Bubba"},
				model.MenuHint{Mnemonic: "b", Description: "Duh"},
				model.MenuHint{Mnemonic: "a", Description: "Blee"},
				model.MenuHint{Mnemonic: "1", Description: "Zorg"},
			},
			e: []int{3, 0, 2, 1},
		},
		"all_strs": {
			hh: model.MenuHints{
				model.MenuHint{Mnemonic: "b", Description: "Bob"},
				model.MenuHint{Mnemonic: "a", Description: "Abby"},
				model.MenuHint{Mnemonic: "c", Description: "Chris"},
			},
			e: []int{1, 0, 2},
		},
		"all_ints": {
			hh: model.MenuHints{
				model.MenuHint{Mnemonic: "3", Description: "Bob"},
				model.MenuHint{Mnemonic: "2", Description: "Abby"},
				model.MenuHint{Mnemonic: "1", Description: "Chris"},
			},
			e: []int{2, 1, 0},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			o := make(model.MenuHints, len(u.hh))
			copy(o, u.hh)
			sort.Sort(u.hh)
			for i, idx := range u.e {
				assert.Equal(t, o[idx], u.hh[i])
			}
		})
	}
}

func TestMenuHintBlank(t *testing.T) {
	uu := map[string]struct {
		hint model.MenuHint
		e    bool
	}{
		"yes": {hint: model.MenuHint{}, e: true},
		"no":  {hint: model.MenuHint{Mnemonic: "a", Description: "blee"}},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.hint.IsBlank())
		})
	}
}
