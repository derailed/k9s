package ui

import (
	"sort"

	"github.com/derailed/k9s/internal/model"
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
		Shared      bool
	}

	// KeyActions tracks mappings between keystrokes and actions.
	KeyActions map[tcell.Key]KeyAction
)

// NewKeyAction returns a new keyboard action.
func NewKeyAction(d string, a ActionHandler, display bool) KeyAction {
	return KeyAction{Description: d, Action: a, Visible: display}
}

// NewSharedKeyAction returns a new shared keyboard action.
func NewSharedKeyAction(d string, a ActionHandler, display bool) KeyAction {
	return KeyAction{Description: d, Action: a, Visible: display, Shared: true}
}

// Add sets up keyboard action listener.
func (a KeyActions) Add(aa KeyActions) {
	for k, v := range aa {
		a[k] = v
	}
}

// Clear remove all actions.
func (a KeyActions) Clear() {
	for k := range a {
		delete(a, k)
	}
}

// Set replace actions with new ones.
func (a KeyActions) Set(aa KeyActions) {
	for k, v := range aa {
		a[k] = v
	}
}

// Delete deletes actions by the given keys.
func (a KeyActions) Delete(kk ...tcell.Key) {
	for _, k := range kk {
		delete(a, k)
	}
}

// Hints returns a collection of hints.
func (a KeyActions) Hints() model.MenuHints {
	kk := make([]int, 0, len(a))
	for k := range a {
		if !a[k].Shared {
			kk = append(kk, int(k))
		}
	}
	sort.Ints(kk)

	hh := make(model.MenuHints, 0, len(kk))
	for _, k := range kk {
		if name, ok := tcell.KeyNames[tcell.Key(k)]; ok {
			hh = append(hh,
				model.MenuHint{
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
