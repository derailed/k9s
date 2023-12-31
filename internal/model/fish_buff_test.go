// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model_test

import (
	"sort"
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestFishAdd(t *testing.T) {
	m := mockSuggestionListener{}

	f := model.NewFishBuff(' ', model.FilterBuffer)
	f.AddListener(&m)
	f.SetSuggestionFn(func(text string) sort.StringSlice {
		return sort.StringSlice{"blee", "brew"}
	})
	f.Add('b')
	f.SetActive(true)

	assert.True(t, m.active)
	assert.Equal(t, 1, m.changeCount)
	assert.Equal(t, 1, m.suggCount)
	assert.Equal(t, "blee", m.suggestion)

	c, ok := f.CurrentSuggestion()
	assert.True(t, ok)
	assert.Equal(t, "blee", c)

	c, ok = f.NextSuggestion()
	assert.True(t, ok)
	assert.Equal(t, "brew", c)

	c, ok = f.PrevSuggestion()
	assert.True(t, ok)
	assert.Equal(t, "blee", c)
}

func TestFishDelete(t *testing.T) {
	m := mockSuggestionListener{}

	f := model.NewFishBuff(' ', model.FilterBuffer)
	f.AddListener(&m)
	f.SetSuggestionFn(func(text string) sort.StringSlice {
		return sort.StringSlice{"blee", "duh"}
	})
	f.Add('a')
	f.Delete()
	f.SetActive(true)

	assert.Equal(t, 2, m.changeCount)
	assert.Equal(t, 3, m.suggCount)
	assert.True(t, m.active)
	assert.Equal(t, "blee", m.suggestion)

	c, ok := f.CurrentSuggestion()
	assert.True(t, ok)
	assert.Equal(t, "blee", c)

	c, ok = f.NextSuggestion()
	assert.True(t, ok)
	assert.Equal(t, "duh", c)

	c, ok = f.PrevSuggestion()
	assert.True(t, ok)
	assert.Equal(t, "blee", c)
}

// Helpers...

type mockSuggestionListener struct {
	changeCount, suggCount int
	suggestion, text       string
	active                 bool
}

func (m *mockSuggestionListener) BufferChanged(_, _ string) {
	m.changeCount++
}

func (m *mockSuggestionListener) BufferCompleted(text, suggest string) {
	if m.suggestion != suggest {
		m.suggCount++
	}
	m.text, m.suggestion = text, suggest
}

func (m *mockSuggestionListener) BufferActive(state bool, kind model.BufferKind) {
	m.active = state
}

func (m *mockSuggestionListener) SuggestionChanged(text, sugg string) {
	m.suggestion = sugg
	m.suggCount++
}
