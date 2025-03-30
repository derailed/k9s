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
	s := config.NewStyles(false)

	assert.Equal(t, config.Color("black"), s.Skin.K9s.Body.BgColor)
	assert.Equal(t, config.Color("cadetblue"), s.Skin.K9s.Body.FgColor)
	assert.Equal(t, config.Color("lightskyblue"), s.Skin.K9s.Frame.Status.NewColor)
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
	s := config.NewStyles(false)
	assert.Nil(t, s.LoadSkin("../../skins/black-and-wtf.yaml"))
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
			s := config.NewStyles(false)
			err := s.LoadSkin(u.f)
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

func TestEmojiForWithValidKeys(t *testing.T) {
	emojiConfig := &config.EmojiConfig{
		K9s: config.EmojiPalette{
			System: config.SystemEmoji{
				LogStreamCancel: "🏁",
				NewVersion:      "⚡️",
				Default:         "📎",
			},
			Prompt: config.PromptEmoji{
				Query:  "🐶",
				Filter: "🐩",
			},
			Status: config.StatusEmoji{
				Info:  "😎",
				Warn:  "😗",
				Error: "😡",
			},
			Xray: config.XrayEmoji{
				Pods:         "🚛",
				Nodes:        "🖥",
				Deployments:  "🪂",
				StatefulSets: "🎎",
				Services:     "💁‍♀️",
				Issue0:       "👍",
				Issue1:       "🔊",
				Issue2:       "☣️",
				Issue3:       "🧨",
			},
		},
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"System Log Stream Cancelled", "system.log_stream_cancelled", "🏁"},
		{"System New Version", "system.new_version", "⚡️"},
		{"Prompt Query", "prompt.query", "🐶"},
		{"Prompt Filter", "prompt.filter", "🐩"},
		{"Status Info", "status.info", "😎"},
		{"Status Warn", "status.warn", "😗"},
		{"Status Error", "status.error", "😡"},
		{"Xray Pods", "xray.pods", "🚛"},
		{"Xray Nodes", "xray.nodes", "🖥"},
		{"Xray Deployments", "xray.deployments", "🪂"},
		{"Xray StatefulSets", "xray.stateful_sets", "🎎"},
		{"Xray Services", "xray.services", "💁‍♀️"},
		{"Xray Issue0", "xray.issue_0", "👍"},
		{"Xray Issue1", "xray.issue_1", "🔊"},
		{"Xray Issue2", "xray.issue_2", "☣️"},
		{"Xray Issue3", "xray.issue_3", "🧨"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := emojiConfig.EmojiFor(tt.key)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestEmojiForWithInvalidKeys(t *testing.T) {
	emojiConfig := &config.EmojiConfig{
		K9s: config.EmojiPalette{
			System: config.SystemEmoji{
				Default:         "📎",
				LogStreamCancel: "🏁",
			},
		},
	}

	tests := []struct {
		name     string
		key      string
		expected string
	}{
		{"Invalid format - no period", "systemstartup", "📎"},
		{"Invalid format - too many parts", "system.startup.extra", "📎"},
		{"Invalid category", "nonexistent.key", "📎"},
		{"Valid category, invalid key", "system.nonexistent", "📎"},
		{"Empty string", "", "📎"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := emojiConfig.EmojiFor(tt.key)
			assert.Equal(t, "📎", result)
		})
	}
}

func TestLoadEmojiInvalid(t *testing.T) {
	uu := map[string]struct {
		f   string
		err string
	}{
		"not-exist": {
			f:   "testdata/emoji/blee.yaml",
			err: "open testdata/emoji/blee.yaml: no such file or directory",
		},
		"invalid": {
			f:   "testdata/emoji/invalid.yaml",
			err: `Invalid type. Expected: object, given: array`,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			s := config.NewStyles(false)
			err := s.LoadEmoji(u.f)
			if err != nil {
				assert.Equal(t, u.err, err.Error())
			}

			// Default emoji values should be preserved
			assert.Equal(t, "🏁", s.Emoji.K9s.System.LogStreamCancel)
			assert.Equal(t, "🐶", s.Emoji.K9s.Prompt.Query)
		})
	}
}

func TestLoadEmojiValid(t *testing.T) {
	s := config.NewStyles(false)

	f := "testdata/emoji/valid.yaml"

	// Check defaults before loading
	assert.Equal(t, "🐶", s.Emoji.K9s.Prompt.Query)
	assert.Equal(t, "📎", s.Emoji.K9s.System.Default)

	// Load custom emoji file
	err := s.LoadEmoji(f)
	assert.Nil(t, err)

	// Check that values are updated
	assert.Equal(t, "🔍", s.Emoji.K9s.Prompt.Query)
	assert.Equal(t, "📝", s.Emoji.K9s.System.Default)
}
