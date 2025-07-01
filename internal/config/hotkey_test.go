// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHotKeyLoad(t *testing.T) {
	h := config.NewHotKeys()
	require.NoError(t, h.LoadHotKeys("testdata/hotkeys/hotkeys.yaml"))
	assert.Len(t, h.HotKey, 1)

	k, ok := h.HotKey["pods"]
	assert.True(t, ok)
	assert.Equal(t, "shift-0", k.ShortCut)
	assert.Equal(t, "Launch pod view", k.Description)
	assert.Equal(t, "pods", k.Command)
	assert.True(t, k.KeepHistory)
}
