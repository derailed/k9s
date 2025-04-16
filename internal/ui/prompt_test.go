// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestCmdNew(t *testing.T) {
	v := ui.NewPrompt(nil, true, config.NewStyles())
	m := model.NewFishBuff(':', model.CommandBuffer)
	v.SetModel(m)
	m.AddListener(v)
	for _, r := range "blee" {
		m.Add(r)
	}

	assert.Equal(t, "\x00> [::b]blee\n", v.GetText(false))
}

func TestCmdUpdate(t *testing.T) {
	m := model.NewFishBuff(':', model.CommandBuffer)
	v := ui.NewPrompt(nil, true, config.NewStyles())
	v.SetModel(m)

	m.AddListener(v)
	m.SetText("blee", "")
	m.Add('!')

	assert.Equal(t, "\x00> [::b]blee!\n", v.GetText(false))
	assert.False(t, v.InCmdMode())
}

func TestCmdMode(t *testing.T) {
	m := model.NewFishBuff(':', model.CommandBuffer)
	v := ui.NewPrompt(&ui.App{}, true, config.NewStyles())
	v.SetModel(m)
	m.AddListener(v)

	for _, f := range []bool{false, true} {
		m.SetActive(f)
		assert.Equal(t, f, v.InCmdMode())
	}
}

func TestPrompt_Deactivate(t *testing.T) {
	m := model.NewFishBuff(':', model.CommandBuffer)
	v := ui.NewPrompt(&ui.App{}, true, config.NewStyles())
	v.SetModel(m)
	m.AddListener(v)

	m.SetActive(true)
	if assert.True(t, v.InCmdMode()) {
		v.Deactivate()
		assert.False(t, v.InCmdMode())
	}
}

// Tests that, when active, the prompt has the appropriate color
func TestPromptColor(t *testing.T) {
	styles := config.NewStyles()
	app := ui.App{}

	// Make sure to have different values to be sure that the prompt color actually changes depending on its type
	assert.NotEqual(t,
		styles.Prompt().Border.DefaultColor.Color(),
		styles.Prompt().Border.CommandColor.Color(),
	)

	testCases := []struct {
		kind          model.BufferKind
		expectedColor tcell.Color
	}{
		// Command prompt case
		{
			kind:          model.CommandBuffer,
			expectedColor: styles.Prompt().Border.CommandColor.Color(),
		},
		// Any other prompt type case
		{
			// Simulate a different type of prompt since no particular constant exists
			kind:          model.CommandBuffer + 1,
			expectedColor: styles.Prompt().Border.DefaultColor.Color(),
		},
	}

	for _, testCase := range testCases {
		m := model.NewFishBuff(':', testCase.kind)
		prompt := ui.NewPrompt(&app, true, styles)

		prompt.SetModel(m)
		m.AddListener(prompt)

		m.SetActive(true)
		assert.Equal(t, testCase.expectedColor, prompt.GetBorderColor())
	}
}

// Tests that, when a change of style occurs, the prompt will have the appropriate color when active
func TestPromptStyleChanged(t *testing.T) {
	app := ui.App{}
	styles := config.NewStyles()
	newStyles := config.NewStyles()
	newStyles.K9s.Prompt.Border = config.PromptBorder{
		DefaultColor: "green",
		CommandColor: "yellow",
	}

	// Check that the prompt won't change the border into the same style
	assert.NotEqual(t, styles.Prompt().Border.CommandColor.Color(), newStyles.Prompt().Border.CommandColor.Color())
	assert.NotEqual(t, styles.Prompt().Border.DefaultColor.Color(), newStyles.Prompt().Border.DefaultColor.Color())

	testCases := []struct {
		kind          model.BufferKind
		expectedColor tcell.Color
	}{
		// Command prompt case
		{
			kind:          model.CommandBuffer,
			expectedColor: newStyles.Prompt().Border.CommandColor.Color(),
		},
		// Any other prompt type case
		{
			// Simulate a different type of prompt since no particular constant exists
			kind:          model.CommandBuffer + 1,
			expectedColor: newStyles.Prompt().Border.DefaultColor.Color(),
		},
	}

	for _, testCase := range testCases {
		m := model.NewFishBuff(':', testCase.kind)
		prompt := ui.NewPrompt(&app, true, styles)

		m.SetActive(true)

		prompt.SetModel(m)
		m.AddListener(prompt)

		prompt.StylesChanged(newStyles)

		m.SetActive(true)
		assert.Equal(t, testCase.expectedColor, prompt.GetBorderColor())
	}
}
