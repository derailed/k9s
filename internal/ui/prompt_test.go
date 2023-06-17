package ui_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/stretchr/testify/assert"
)

func TestCmdNew(t *testing.T) {
	v := ui.NewPrompt(nil, true, config.NewStyles())
	model := model.NewFishBuff(':', model.CommandBuffer)
	v.SetModel(model)
	model.AddListener(v)
	for _, r := range "blee" {
		model.Add(r)
	}

	assert.Equal(t, "\x00> [::b]blee\n", v.GetText(false))
}

func TestCmdUpdate(t *testing.T) {
	model := model.NewFishBuff(':', model.CommandBuffer)
	v := ui.NewPrompt(nil, true, config.NewStyles())
	v.SetModel(model)

	model.AddListener(v)
	model.SetText("blee", "")
	model.Add('!')

	assert.Equal(t, "\x00> [::b]blee!\n", v.GetText(false))
	assert.False(t, v.InCmdMode())
}

func TestCmdMode(t *testing.T) {
	model := model.NewFishBuff(':', model.CommandBuffer)
	v := ui.NewPrompt(&ui.App{}, true, config.NewStyles())
	v.SetModel(model)
	model.AddListener(v)

	for _, f := range []bool{false, true} {
		model.SetActive(f)
		assert.Equal(t, f, v.InCmdMode())
	}
}

func TestPromptStylesChanged(t *testing.T) {
	style := config.NewStyles()
	prompt := ui.NewPrompt(nil, true, style)

	// Check that the style is respected when the prompt is created
	assert.Equal(t, prompt.TextView.Box.GetBorderColor(), style.Frame().Border.FgColor.Color().TrueColor())
	assert.Equal(t, prompt.TextView.Box.GetBackgroundColor(), style.K9s.Prompt.BgColor.Color())

	// Create a new style with a different border and background color
	newStyle := config.NewStyles()
	newStyle.K9s.Frame.Border = config.Border{
		FgColor:    "red",
		FocusColor: "red",
	}
	newStyle.K9s.Prompt = config.Prompt {
		FgColor:      "red",
		BgColor:      "red",
		SuggestColor: "red",
	}

	// Make sure the new style is different from the first one
	assert.NotEqual(t, style.Frame().Border.FgColor.Color().TrueColor(), newStyle.Frame().Border.FgColor.Color().TrueColor())
	assert.NotEqual(t, style.K9s.Prompt.BgColor.Color(), newStyle.K9s.Prompt.BgColor.Color())

	prompt.StylesChanged(newStyle)

	assert.Equal(t, prompt.TextView.Box.GetBorderColor(), newStyle.Frame().Border.FgColor.Color().TrueColor())
	assert.Equal(t, prompt.TextView.Box.GetBackgroundColor(), newStyle.K9s.Prompt.BgColor.Color())
}
