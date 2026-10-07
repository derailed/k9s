// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Enter on a PC keyboard sends LF (Ctrl-J, key code 10) instead of CR.
// The app-level input capture must translate it so every view treats it
// like the Mac Return key.
func TestAppKeyboardTranslatesLFToEnter(t *testing.T) {
	app := NewApp(mock.NewMockConfig(t))

	uu := map[string]struct {
		key tcell.Key
		e   tcell.Key
	}{
		"lf":    {key: tcell.KeyLF, e: tcell.KeyEnter},
		"enter": {key: tcell.KeyEnter, e: tcell.KeyEnter},
		"other": {key: tcell.KeyDown, e: tcell.KeyDown},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			evt := tcell.NewEventKey(u.key, 0, tcell.ModNone)
			res := app.keyboard(evt)
			require.NotNil(t, res)
			assert.Equal(t, u.e, res.Key())
		})
	}
}
