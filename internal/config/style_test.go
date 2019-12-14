package config

import (
	"testing"

	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
)

func TestSkinNone(t *testing.T) {
	s, err := NewStyles("test_assets/empty_skin.yml")
	assert.Nil(t, err)

	s.Update()

	assert.Equal(t, "cadetblue", s.Body().FgColor)
	assert.Equal(t, "black", s.Body().BgColor)
	assert.Equal(t, "black", s.GetTable().BgColor)
	assert.Equal(t, tcell.ColorCadetBlue, s.FgColor())
	assert.Equal(t, tcell.ColorBlack, s.BgColor())
	assert.Equal(t, tcell.ColorBlack, tview.Styles.PrimitiveBackgroundColor)
	assert.Equal(t, tcell.ColorPink, AsColor("blah"))
	assert.Equal(t, tcell.ColorWhite, AsColor("white"))
}

func TestSkin(t *testing.T) {
	s, err := NewStyles("test_assets/black_and_wtf.yml")
	assert.Nil(t, err)

	s.Update()

	assert.Equal(t, "white", s.Body().FgColor)
	assert.Equal(t, "black", s.Body().BgColor)
	assert.Equal(t, "black", s.GetTable().BgColor)
	assert.Equal(t, tcell.ColorWhite, s.FgColor())
	assert.Equal(t, tcell.ColorBlack, s.BgColor())
	assert.Equal(t, tcell.ColorBlack, tview.Styles.PrimitiveBackgroundColor)
	assert.Equal(t, tcell.ColorPink, AsColor("blah"))
	assert.Equal(t, tcell.ColorWhite, AsColor("white"))
}

func TestSkinNotExits(t *testing.T) {
	_, err := NewStyles("test_assets/blee.yml")
	assert.NotNil(t, err)
}

func TestSkinBoarked(t *testing.T) {
	_, err := NewStyles("test_assets/skin_boarked.yml")
	assert.NotNil(t, err)
}
