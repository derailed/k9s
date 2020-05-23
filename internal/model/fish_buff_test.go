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
		return sort.StringSlice{"blee", "duh"}
	})
	f.Add('a')
	f.SetActive(true)

	assert.Equal(t, 1, m.buff)
	assert.Equal(t, 1, m.sugg)
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

	assert.Equal(t, 2, m.buff)
	assert.Equal(t, 2, m.sugg)
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
	buff, sugg int
	suggestion string
	active     bool
}

func (m *mockSuggestionListener) BufferChanged(s string) {
	m.buff++
}

func (m *mockSuggestionListener) BufferActive(state bool, kind model.BufferKind) {
	m.active = state
}

func (m *mockSuggestionListener) SuggestionChanged(text, sugg string) {
	m.suggestion = sugg
	m.sugg++
}
