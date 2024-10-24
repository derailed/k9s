// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestNewStyle(t *testing.T) {
	s := config.NewStyles()

	assert.Equal(t, config.Color("black"), s.K9s.Body.BgColor)
	assert.Equal(t, config.Color("cadetblue"), s.K9s.Body.FgColor)
	assert.Equal(t, config.Color("lightskyblue"), s.K9s.Frame.Status.NewColor)
}

func TestColor(t *testing.T) {
	uu := map[string]tcell.Color{
		"blah":    tcell.ColorDefault,
		"blue":    tcell.ColorBlue.TrueColor(),
		"#ffffff": tcell.NewHexColor(16777215),
		"#ff0000": tcell.NewHexColor(16711680),
	}

	for k := range uu {
		c, u := k, uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u, config.NewColor(c).Color())
		})
	}
}

func TestSkinHappy(t *testing.T) {
	s := config.NewStyles()
	assert.Nil(t, s.Load("../../skins/black-and-wtf.yaml"))
	s.Update()

	assert.Equal(t, "#ffffff", s.Body().FgColor.String())
	assert.Equal(t, "#000000", s.Body().BgColor.String())
	assert.Equal(t, "#000000", s.Table().BgColor.String())
	assert.Equal(t, tcell.ColorWhite.TrueColor(), s.FgColor())
	assert.Equal(t, tcell.ColorBlack.TrueColor(), s.BgColor())
	assert.Equal(t, tcell.ColorBlack.TrueColor(), tview.Styles.PrimitiveBackgroundColor)
}

func TestSkinLoad(t *testing.T) {
	uu := map[string]struct {
		f   string
		err string
	}{
		"not-exist": {
			f:   "testdata/skins/blee.yaml",
			err: "open testdata/skins/blee.yaml: no such file or directory",
		},
		"toast": {
			f: "testdata/skins/boarked.yaml",
			err: `Additional property bgColor is not allowed
Additional property fgColor is not allowed
Additional property logoColor is not allowed
Invalid type. Expected: object, given: array`,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			s := config.NewStyles()
			err := s.Load(u.f)
			if err != nil {
				assert.Equal(t, u.err, err.Error())
			}
			assert.Equal(t, "#5f9ea0", s.Body().FgColor.String())
			assert.Equal(t, "#000000", s.Body().BgColor.String())
			assert.Equal(t, "#000000", s.Table().BgColor.String())
			assert.Equal(t, tcell.ColorCadetBlue.TrueColor(), s.FgColor())
			assert.Equal(t, tcell.ColorBlack.TrueColor(), s.BgColor())
			assert.Equal(t, tcell.ColorBlack.TrueColor(), tview.Styles.PrimitiveBackgroundColor)
		})
	}
}
