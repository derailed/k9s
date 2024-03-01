// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"sort"
	"sync"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/tcell/v2"
	"github.com/rs/zerolog/log"
)

type (
	// RangeFn represents a range iteration callback.
	RangeFn func(tcell.Key, KeyAction)

	// ActionHandler handles a keyboard command.
	ActionHandler func(*tcell.EventKey) *tcell.EventKey

	// ActionOpts tracks various action options.
	ActionOpts struct {
		Visible   bool
		Shared    bool
		Plugin    bool
		HotKey    bool
		Dangerous bool
	}

	// KeyAction represents a keyboard action.
	KeyAction struct {
		Description string
		Action      ActionHandler
		Opts        ActionOpts
	}

	// KeyMap tracks key to action mappings.
	KeyMap map[tcell.Key]KeyAction

	// KeyActions tracks mappings between keystrokes and actions.
	KeyActions struct {
		actions KeyMap
		mx      sync.RWMutex
	}
)

// NewKeyAction returns a new keyboard action.
func NewKeyAction(d string, a ActionHandler, visible bool) KeyAction {
	return NewKeyActionWithOpts(d, a, ActionOpts{
		Visible: visible,
	})
}

// NewSharedKeyAction returns a new shared keyboard action.
func NewSharedKeyAction(d string, a ActionHandler, visible bool) KeyAction {
	return NewKeyActionWithOpts(d, a, ActionOpts{
		Visible: visible,
		Shared:  true,
	})
}

// NewKeyActionWithOpts returns a new keyboard action.
func NewKeyActionWithOpts(d string, a ActionHandler, opts ActionOpts) KeyAction {
	return KeyAction{
		Description: d,
		Action:      a,
		Opts:        opts,
	}
}

// NewKeyActions returns a new instance.
func NewKeyActions() *KeyActions {
	return &KeyActions{
		actions: make(map[tcell.Key]KeyAction),
	}
}

// NewKeyActionsFromMap construct actions from key map.
func NewKeyActionsFromMap(mm KeyMap) *KeyActions {
	return &KeyActions{actions: mm}
}

// Get fetches an action given a key.
func (a *KeyActions) Get(key tcell.Key) (KeyAction, bool) {
	a.mx.RLock()
	defer a.mx.RUnlock()

	v, ok := a.actions[key]

	return v, ok
}

// Len returns action mapping count.
func (a *KeyActions) Len() int {
	a.mx.RLock()
	defer a.mx.RUnlock()

	return len(a.actions)
}

// Reset clears out actions.
func (a *KeyActions) Reset(aa *KeyActions) {
	a.Clear()
	a.Merge(aa)
}

// Range ranges over all actions and triggers a given function.
func (a *KeyActions) Range(f RangeFn) {
	var km KeyMap
	a.mx.RLock()
	{
		km = a.actions
	}
	a.mx.RUnlock()

	for k, v := range km {
		f(k, v)
	}
}

// Add adds a new key action.
func (a *KeyActions) Add(k tcell.Key, ka KeyAction) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.actions[k] = ka
}

// Bulk bulk insert key mappings.
func (a *KeyActions) Bulk(aa KeyMap) {
	a.mx.Lock()
	defer a.mx.Unlock()

	for k, v := range aa {
		a.actions[k] = v
	}
}

// Merge merges given actions into existing set.
func (a *KeyActions) Merge(aa *KeyActions) {
	a.mx.Lock()
	defer a.mx.Unlock()

	for k, v := range aa.actions {
		a.actions[k] = v
	}
}

// Clear remove all actions.
func (a *KeyActions) Clear() {
	a.mx.Lock()
	defer a.mx.Unlock()

	for k := range a.actions {
		delete(a.actions, k)
	}
}

// ClearDanger remove all dangerous actions.
func (a *KeyActions) ClearDanger() {
	a.mx.Lock()
	defer a.mx.Unlock()

	for k, v := range a.actions {
		if v.Opts.Dangerous {
			delete(a.actions, k)
		}
	}
}

// Set replace actions with new ones.
func (a *KeyActions) Set(aa *KeyActions) {
	a.mx.Lock()
	defer a.mx.Unlock()

	for k, v := range aa.actions {
		a.actions[k] = v
	}
}

// Delete deletes actions by the given keys.
func (a *KeyActions) Delete(kk ...tcell.Key) {
	a.mx.Lock()
	defer a.mx.Unlock()

	for _, k := range kk {
		delete(a.actions, k)
	}
}

// Hints returns a collection of hints.
func (a *KeyActions) Hints() model.MenuHints {
	a.mx.RLock()
	defer a.mx.RUnlock()

	kk := make([]int, 0, len(a.actions))
	for k := range a.actions {
		if !a.actions[k].Opts.Shared {
			kk = append(kk, int(k))
		}
	}
	sort.Ints(kk)

	hh := make(model.MenuHints, 0, len(kk))
	for _, k := range kk {
		if name, ok := tcell.KeyNames[tcell.Key(int16(k))]; ok {
			hh = append(hh,
				model.MenuHint{
					Mnemonic:    name,
					Description: a.actions[tcell.Key(k)].Description,
					Visible:     a.actions[tcell.Key(k)].Opts.Visible,
				},
			)
		} else {
			log.Error().Msgf("Unable to locate KeyName for %#v", k)
		}
	}

	return hh
}
