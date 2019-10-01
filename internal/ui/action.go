package ui

import (
	"sort"

	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

type (
	// ActionHandler handles a keyboard command.
	ActionHandler func(*tcell.EventKey) *tcell.EventKey

	// KeyAction represents a keyboard action.
	KeyAction struct {
		Description string
		Action      ActionHandler
		Visible     bool
	}

	// KeyActions tracks mappings between keystrokes and actions.
	KeyActions map[tcell.Key]KeyAction
)

// NewKeyAction returns a new keyboard action.
func NewKeyAction(d string, a ActionHandler, display bool) KeyAction {
	return KeyAction{Description: d, Action: a, Visible: display}
}

// Hints returns a collection of hints.
func (a KeyActions) Hints() Hints {
	kk := make([]int, 0, len(a))
	for k := range a {
		kk = append(kk, int(k))
	}
	sort.Ints(kk)

	hh := make(Hints, 0, len(kk))
	for _, k := range kk {
		if name, ok := tcell.KeyNames[tcell.Key(k)]; ok {
			hh = append(hh,
				Hint{
					Mnemonic:    name,
					Description: a[tcell.Key(k)].Description,
					Visible:     a[tcell.Key(k)].Visible,
				},
			)
		} else {
			log.Error().Msgf("Unable to locate KeyName for %#v", string(k))
		}
	}
	return hh
}
