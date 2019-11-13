package ui_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
)

func TestNewMenu(t *testing.T) {
	defaults, _ := config.NewStyles("")
	v := ui.NewMenu(defaults)
	v.HydrateMenu(model.MenuHints{
		{Mnemonic: "a", Description: "bleeA", Visible: true},
		{Mnemonic: "b", Description: "bleeB", Visible: true},
		{Mnemonic: "0", Description: "zero", Visible: true},
	})

	assert.Equal(t, " [fuchsia:black:b]<0> [white:black:d]zero ", v.GetCell(0, 0).Text)
	assert.Equal(t, " [dodgerblue:black:b]<a> [white:black:d]bleeA ", v.GetCell(0, 1).Text)
	assert.Equal(t, " [dodgerblue:black:b]<b> [white:black:d]bleeB ", v.GetCell(1, 1).Text)
}

func TestActionHints(t *testing.T) {
	uu := map[string]struct {
		aa ui.KeyActions
		e  model.MenuHints
	}{
		"a": {
			aa: ui.KeyActions{
				ui.KeyB:            ui.NewKeyAction("bleeB", nil, true),
				ui.KeyA:            ui.NewKeyAction("bleeA", nil, true),
				tcell.Key(ui.Key0): ui.NewKeyAction("zero", nil, true),
				tcell.Key(ui.Key1): ui.NewKeyAction("one", nil, false),
			},
			e: model.MenuHints{
				{Mnemonic: "0", Description: "zero", Visible: true},
				{Mnemonic: "1", Description: "one", Visible: false},
				{Mnemonic: "a", Description: "bleeA", Visible: true},
				{Mnemonic: "b", Description: "bleeB", Visible: true},
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.aa.Hints())
		})
	}
}
