package ui

import (
	"testing"

	"github.com/gdamore/tcell"
	"github.com/stretchr/testify/assert"
)

func TestFlashEmoji(t *testing.T) {
	uu := []struct {
		level FlashLevel
		emoji string
	}{
		{FlashWarn, emoDoh},
		{FlashErr, emoRed},
		{FlashFatal, emoDead},
		{FlashInfo, emoHappy},
	}

	for _, u := range uu {
		assert.Equal(t, u.emoji, flashEmoji(u.level))
	}
}

func TestFlashColor(t *testing.T) {
	uu := []struct {
		level FlashLevel
		color tcell.Color
	}{
		{FlashWarn, tcell.ColorOrange},
		{FlashErr, tcell.ColorOrangeRed},
		{FlashFatal, tcell.ColorFuchsia},
		{FlashInfo, tcell.ColorNavajoWhite},
	}

	for _, u := range uu {
		assert.Equal(t, u.color, flashColor(u.level))
	}

}
