package config_test

import (
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestColor(t *testing.T) {
	uu := map[string]tcell.Color{
		"blah":    tcell.ColorDefault,
		"blue":    tcell.ColorBlue,
		"#ffffff": tcell.NewHexColor(16777215),
		"#ff0000": tcell.NewHexColor(16711680),
	}

	for k := range uu {
		c, u := k, uu[k]
		t.Run(k, func(t *testing.T) {
			fmt.Printf("%#v\n", config.NewColor(c).Color().Hex())
			assert.Equal(t, u, config.NewColor(c).Color())
		})
	}
}

func TestSkinNone(t *testing.T) {
	s := config.NewStyles()
	assert.Nil(t, s.Load("testdata/empty_skin.yml"))
	s.Update()

	assert.Equal(t, "cadetblue", s.Body().FgColor.String())
	assert.Equal(t, "black", s.Body().BgColor.String())
	assert.Equal(t, "black", s.Table().BgColor.String())
	assert.Equal(t, tcell.ColorCadetBlue, s.FgColor())
	assert.Equal(t, tcell.ColorBlack, s.BgColor())
	assert.Equal(t, tcell.ColorBlack, tview.Styles.PrimitiveBackgroundColor)
}

func TestSkin(t *testing.T) {
	s := config.NewStyles()
	assert.Nil(t, s.Load("testdata/black_and_wtf.yml"))
	s.Update()

	assert.Equal(t, "white", s.Body().FgColor.String())
	assert.Equal(t, "black", s.Body().BgColor.String())
	assert.Equal(t, "black", s.Table().BgColor.String())
	assert.Equal(t, tcell.ColorWhite, s.FgColor())
	assert.Equal(t, tcell.ColorBlack, s.BgColor())
	assert.Equal(t, tcell.ColorBlack, tview.Styles.PrimitiveBackgroundColor)
}

func TestSkinNotExits(t *testing.T) {
	s := config.NewStyles()
	assert.NotNil(t, s.Load("testdata/blee.yml"))
}

func TestSkinBoarked(t *testing.T) {
	s := config.NewStyles()
	assert.NotNil(t, s.Load("testdata/skin_boarked.yml"))
}
