// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClipboardMode(t *testing.T) {
	uu := map[string]struct {
		env string
		e   string
	}{
		"empty":   {env: "", e: clipboardModeAuto},
		"auto":    {env: "auto", e: clipboardModeAuto},
		"native":  {env: "native", e: clipboardModeNative},
		"osc52":   {env: "osc52", e: clipboardModeOSC52},
		"upper":   {env: "OSC52", e: clipboardModeOSC52},
		"unknown": {env: "toast", e: clipboardModeAuto},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			t.Setenv(clipboardModeEnv, u.env)
			assert.Equal(t, u.e, clipboardMode())
		})
	}
}

func TestOSC52MaxEncodedLen(t *testing.T) {
	uu := map[string]struct {
		env string
		e   int
	}{
		"empty":    {env: "", e: defaultOSC52MaxEncodedLen},
		"valid":    {env: "100", e: 100},
		"zero":     {env: "0", e: defaultOSC52MaxEncodedLen},
		"negative": {env: "-1", e: defaultOSC52MaxEncodedLen},
		"invalid":  {env: "abc", e: defaultOSC52MaxEncodedLen},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			t.Setenv(osc52MaxEnv, u.env)
			assert.Equal(t, u.e, osc52MaxEncodedLen())
		})
	}
}

func TestOSC52Sequence(t *testing.T) {
	const encoded = "SGVsbG8="

	uu := map[string]struct {
		tmux, screen bool
		e            string
	}{
		"plain":  {e: "\033]52;c;SGVsbG8=\a"},
		"tmux":   {tmux: true, e: "\033Ptmux;\033\033]52;c;SGVsbG8=\a\033\\"},
		"screen": {screen: true, e: "\033P\033]52;c;SGVsbG8=\a\033\\"},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, osc52Sequence(encoded, u.tmux, u.screen))
		})
	}
}

func TestWriteOSC52Unavailable(t *testing.T) {
	t.Setenv(termEnv, dumbTerm)

	err := writeOSC52("hello")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "osc52 clipboard unavailable")
}
