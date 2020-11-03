package model_test

import (
	"sort"
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestFishExact(t *testing.T) {
	m := mockSuggestionListener{}

	f := model.NewFishBuff(' ', model.FilterBuffer)
	f.AddListener(&m)
	f.SetSuggestionFn(func(text string) sort.StringSlice {
		return sort.StringSlice{"lee"}
	})
	f.Add('b')
	f.SetActive(true)

	assert.True(t, m.active)
	assert.Equal(t, 1, m.buff)
	assert.Equal(t, 0, m.sugg)
	assert.Equal(t, "blee", m.text)
}

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
	assert.Equal(t, 1, m.buff)
	assert.Equal(t, 1, m.sugg)
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
	buff, sugg       int
	suggestion, text string
	active           bool
}

func (m *mockSuggestionListener) BufferChanged(s string) {
	m.buff++
}

func (m *mockSuggestionListener) BufferCompleted(s string) {
	m.text = s
}

func (m *mockSuggestionListener) BufferActive(state bool, kind model.BufferKind) {
	m.active = state
}

func (m *mockSuggestionListener) SuggestionChanged(text, sugg string) {
	m.suggestion = sugg
	m.sugg++
}
