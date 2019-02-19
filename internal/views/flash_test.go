package views

import (
	"testing"

	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
)

func TestFlashEmoji(t *testing.T) {
	uu := []struct {
		level flashLevel
		emoji string
	}{
		{flashWarn, emoDoh},
		{flashErr, emoRed},
		{flashFatal, emoDead},
		{flashInfo, emoHappy},
	}

	for _, u := range uu {
		assert.Equal(t, u.emoji, flashEmoji(u.level))
	}
}

func TestFlashColor(t *testing.T) {
	uu := []struct {
		level flashLevel
		color tcell.Color
	}{
		{flashWarn, tcell.ColorOrange},
		{flashErr, tcell.ColorOrangeRed},
		{flashFatal, tcell.ColorFuchsia},
		{flashInfo, tcell.ColorNavajoWhite},
	}

	for _, u := range uu {
		assert.Equal(t, u.color, flashColor(u.level))
	}

}
