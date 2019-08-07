package ui

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
)

func TestNewMenuView(t *testing.T) {
	defaults, _ := config.NewStyles("")
	v := NewMenuView(defaults)
	v.HydrateMenu(Hints{
		{"a", "bleeA"},
		{"b", "bleeB"},
		{"0", "zero"},
	})

	assert.Equal(t, " [fuchsia:black:b]<0> [white:black:d]zero ", v.GetCell(0, 0).Text)
	assert.Equal(t, " [dodgerblue:black:b]<a> [white:black:d]bleeA ", v.GetCell(0, 1).Text)
	assert.Equal(t, " [dodgerblue:black:b]<b> [white:black:d]bleeB ", v.GetCell(1, 1).Text)
}

func TestKeyActions(t *testing.T) {
	uu := map[string]struct {
		aa KeyActions
		e  Hints
	}{
		"a": {
			aa: KeyActions{
				KeyB:            NewKeyAction("bleeB", nil, true),
				KeyA:            NewKeyAction("bleeA", nil, true),
				tcell.Key(Key0): NewKeyAction("zero", nil, true),
				tcell.Key(Key1): NewKeyAction("one", nil, false),
			},
			e: Hints{
				{"0", "zero"},
				{"a", "bleeA"},
				{"b", "bleeB"},
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.aa.Hints())
		})
	}
}
