// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestColors(t *testing.T) {
	uu := map[string]struct {
		cc []string
		ee []tcell.Color
	}{
		"empty": {
			ee: []tcell.Color{},
		},
		"default": {
			cc: []string{"default"},
			ee: []tcell.Color{tcell.ColorDefault},
		},
		"multi": {
			cc: []string{
				"default",
				"transparent",
				"blue",
				"green",
			},
			ee: []tcell.Color{
				tcell.ColorDefault,
				tcell.ColorDefault,
				tcell.ColorBlue.TrueColor(),
				tcell.ColorGreen.TrueColor(),
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			cc := make(config.Colors, 0, len(u.cc))
			for _, c := range u.cc {
				cc = append(cc, config.NewColor(c))
			}
			assert.Equal(t, u.ee, cc.Colors())
		})
	}
}

func TestColorString(t *testing.T) {
	uu := map[string]struct {
		c string
		e string
	}{
		"empty": {
			e: "-",
		},
		"default": {
			c: "default",
			e: "-",
		},
		"transparent": {
			c: "-",
			e: "-",
		},
		"blue": {
			c: "blue",
			e: "#0000ff",
		},
		"lightgray": {
			c: "lightgray",
			e: "#d3d3d3",
		},
		"hex": {
			c: "#00ff00",
			e: "#00ff00",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := config.NewColor(u.c)
			assert.Equal(t, u.e, c.String())
		})
	}
}

func TestColorToColor(t *testing.T) {
	uu := map[string]struct {
		c string
		e tcell.Color
	}{
		"default": {
			c: "default",
			e: tcell.ColorDefault,
		},
		"transparent": {
			c: "-",
			e: tcell.ColorDefault,
		},
		"aqua": {
			c: "aqua",
			e: tcell.ColorAqua.TrueColor(),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			c := config.NewColor(u.c)
			assert.Equal(t, u.e, c.Color())
		})
	}
}
