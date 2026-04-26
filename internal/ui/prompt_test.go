// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"sort"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func TestCmdNew(t *testing.T) {
	uu := map[string]struct {
		mode   rune
		kind   model.BufferKind
		noIcon bool
		e      string
	}{
		"cmd": {
			mode:   ':',
			noIcon: true,
			kind:   model.CommandBuffer,
			e:      " > [::b]blee\n",
		},

		"cmd-ic": {
			mode: ':',
			kind: model.CommandBuffer,
			e:    "🐶> [::b]blee\n",
		},

		"search": {
			mode:   '/',
			kind:   model.FilterBuffer,
			noIcon: true,
			e:      " / [::b]blee\n",
		},

		"search-ic": {
			mode: '/',
			kind: model.FilterBuffer,
			e:    "🐩/ [::b]blee\n",
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			v := ui.NewPrompt(nil, u.noIcon, config.NewStyles())
			m := model.NewFishBuff(u.mode, u.kind)
			v.SetModel(m)
			m.AddListener(v)
			for _, r := range "blee" {
				m.Add(r)
			}
			m.SetActive(true)
			assert.Equal(t, u.e, v.GetText(false))
		})
	}
}

func TestCmdUpdate(t *testing.T) {
	m := model.NewFishBuff(':', model.CommandBuffer)
	v := ui.NewPrompt(nil, true, config.NewStyles())
	v.SetModel(m)

	m.AddListener(v)
	m.SetText("blee", "", true)
	m.Add('!')

	assert.Equal(t, "\x00\x00 [::b]blee!\n", v.GetText(false))
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

func TestPromptCommandSuggestionsUseDropdown(t *testing.T) {
	m := model.NewFishBuff(':', model.CommandBuffer)
	m.SetSuggestionFn(func(string) sort.StringSlice {
		return sort.StringSlice{"line", "lines"}
	})
	v := ui.NewPrompt(&ui.App{}, true, config.NewStyles())
	v.SetModel(m)
	m.AddListener(v)

	m.SetActive(true)
	v.SendStrokes("pipe")

	assert.Equal(t, " > [::b]pipe\n", v.GetText(false))
	assert.True(t, v.SuggestionDropdown().IsActive())
	assert.Equal(t, []string{"pipeline", "pipelines"}, v.SuggestionDropdown().Items())
	assert.Equal(t, 0, v.SuggestionDropdown().SelectedIndex())
}

func TestPromptFilterSuggestionsStayInline(t *testing.T) {
	m := model.NewFishBuff('/', model.FilterBuffer)
	m.SetSuggestionFn(func(string) sort.StringSlice {
		return sort.StringSlice{"line"}
	})
	v := ui.NewPrompt(&ui.App{}, true, config.NewStyles())
	v.SetModel(m)
	m.AddListener(v)

	m.SetActive(true)
	v.SendStrokes("pipe")

	assert.Contains(t, v.GetText(false), "line")
	assert.False(t, v.SuggestionDropdown().IsActive())
}

func TestPromptCommandNoSuggestionsHidesDropdown(t *testing.T) {
	m := model.NewFishBuff(':', model.CommandBuffer)
	m.SetSuggestionFn(func(string) sort.StringSlice {
		return nil
	})
	v := ui.NewPrompt(&ui.App{}, true, config.NewStyles())
	v.SetModel(m)
	m.AddListener(v)

	m.SetActive(true)
	v.SendStrokes("pipe")

	assert.False(t, v.SuggestionDropdown().IsActive())
}

func TestPromptCommandSuggestionSelectionTracksArrows(t *testing.T) {
	m := model.NewFishBuff(':', model.CommandBuffer)
	m.SetSuggestionFn(func(string) sort.StringSlice {
		return sort.StringSlice{"line", "linerun", "lines"}
	})
	v := ui.NewPrompt(&ui.App{}, true, config.NewStyles())
	v.SetModel(m)
	m.AddListener(v)

	m.SetActive(true)
	v.SendStrokes("pipe")
	v.SendKey(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))

	assert.Equal(t, " > [::b]pipe\n", v.GetText(false))
	assert.Equal(t, 1, v.SuggestionDropdown().SelectedIndex())
	assert.Equal(t, []string{"pipeline", "pipelinerun", "pipelines"}, v.SuggestionDropdown().Items())
}

func TestPromptCommandEnterAcceptsSelectedSuggestion(t *testing.T) {
	m := model.NewFishBuff(':', model.CommandBuffer)
	m.SetSuggestionFn(func(string) sort.StringSlice {
		return sort.StringSlice{"line", "linerun"}
	})
	v := ui.NewPrompt(&ui.App{}, true, config.NewStyles())
	v.SetModel(m)
	m.AddListener(v)

	m.SetActive(true)
	v.SendStrokes("pipe")
	v.SendKey(tcell.NewEventKey(tcell.KeyDown, 0, tcell.ModNone))
	v.SendKey(tcell.NewEventKey(tcell.KeyEnter, 0, tcell.ModNone))

	assert.Equal(t, "pipelinerun", m.GetText())
	assert.False(t, v.SuggestionDropdown().IsActive())
}

func TestPromptCommandUpMovesToPreviousSuggestion(t *testing.T) {
	m := model.NewFishBuff(':', model.CommandBuffer)
	m.SetSuggestionFn(func(string) sort.StringSlice {
		return sort.StringSlice{"line", "linerun", "lines"}
	})
	v := ui.NewPrompt(&ui.App{}, true, config.NewStyles())
	v.SetModel(m)
	m.AddListener(v)

	m.SetActive(true)
	v.SendStrokes("pipe")
	v.SendKey(tcell.NewEventKey(tcell.KeyUp, 0, tcell.ModNone))

	assert.Equal(t, 2, v.SuggestionDropdown().SelectedIndex())
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
