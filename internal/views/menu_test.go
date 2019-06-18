package views

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
)

func TestNewMenuView(t *testing.T) {
	defaults, _ := config.NewStyles("")
	v := newMenuView(defaults)
	v.populateMenu(hints{
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
		aa keyActions
		e  hints
	}{
		"a": {
			aa: keyActions{
				KeyB:            newKeyAction("bleeB", nil, true),
				KeyA:            newKeyAction("bleeA", nil, true),
				tcell.Key(Key0): newKeyAction("zero", nil, true),
				tcell.Key(Key1): newKeyAction("one", nil, false),
			},
			e: hints{
				{"0", "zero"},
				{"a", "bleeA"},
				{"b", "bleeB"},
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.aa.toHints())
		})
	}
}
