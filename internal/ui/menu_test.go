// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestNewMenu(t *testing.T) {
	v := ui.NewMenu(config.NewStyles())
	v.HydrateMenu(model.MenuHints{
		{Mnemonic: "a", Description: "bleeA", Visible: true},
		{Mnemonic: "b", Description: "bleeB", Visible: true},
		{Mnemonic: "0", Description: "zero", Visible: true},
	})

	assert.Equal(t, " [#ff00ff:-:b]<0> [#ffffff:-:d]zero ", v.GetCell(0, 0).Text)
	assert.Equal(t, " [#1e90ff:-:b]<a> [#ffffff:-:d]bleeA ", v.GetCell(0, 1).Text)
	assert.Equal(t, " [#1e90ff:-:b]<b> [#ffffff:-:d]bleeB ", v.GetCell(1, 1).Text)
}

func TestActionHints(t *testing.T) {
	uu := map[string]struct {
		aa *ui.KeyActions
		e  model.MenuHints
	}{
		"a": {
			aa: ui.NewKeyActionsFromMap(ui.KeyMap{
				ui.KeyB: ui.NewKeyAction("bleeB", nil, true),
				ui.KeyA: ui.NewKeyAction("bleeA", nil, true),
				ui.Key0: ui.NewKeyAction("zero", nil, true),
				ui.Key1: ui.NewKeyAction("one", nil, false),
			}),
			e: model.MenuHints{
				{Mnemonic: "0", Description: "zero", Visible: true},
				{Mnemonic: "1", Description: "one", Visible: false},
				{Mnemonic: "a", Description: "bleeA", Visible: true},
				{Mnemonic: "b", Description: "bleeB", Visible: true},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.aa.Hints())
		})
	}
}
