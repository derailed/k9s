// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package color_test

import (
	"testing"

	"github.com/derailed/k9s/internal/color"
	"github.com/stretchr/testify/assert"
)

func TestColorize(t *testing.T) {
	uu := map[string]struct {
		s string
		c color.Paint
		e string
	}{
		"white":   {"blee", color.LightGray, "\x1b[37mblee\x1b[0m"},
		"black":   {"blee", color.Black, "\x1b[30mblee\x1b[0m"},
		"default": {"blee", 0, "blee"},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, color.Colorize(u.s, u.c))
		})
	}
}

func TestHighlight(t *testing.T) {
	uu := map[string]struct {
		text    []byte
		indices []int
		color   int
		e       string
	}{
		"white": {
			text:    []byte("the brown fox"),
			color:   209,
			indices: []int{4, 5, 6, 7, 8},
			e:       "the \x1b[38;5;209mb\x1b[0m\x1b[38;5;209mr\x1b[0m\x1b[38;5;209mo\x1b[0m\x1b[38;5;209mw\x1b[0m\x1b[38;5;209mn\x1b[0m fox",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, string(color.Highlight([]byte(u.text), u.indices, u.color)))
		})
	}
}
