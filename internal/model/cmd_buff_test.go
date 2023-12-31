// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/stretchr/testify/assert"
)

type testListener struct {
	text, suggestion string
	act              int
	inact            int
}

func (l *testListener) BufferChanged(t, s string) {
	l.text, l.suggestion = t, s
}

func (l *testListener) BufferCompleted(t, s string) {
	l.text, l.suggestion = t, s
}

func (l *testListener) BufferActive(s bool, _ model.BufferKind) {
	if s {
		l.act++
		return
	}
	l.inact++
}

func TestCmdBuffActivate(t *testing.T) {
	b, l := model.NewCmdBuff('>', model.CommandBuffer), testListener{}
	b.AddListener(&l)

	b.SetActive(true)
	assert.Equal(t, 1, l.act)
	assert.Equal(t, 0, l.inact)
	assert.True(t, b.IsActive())
}

func TestCmdBuffDeactivate(t *testing.T) {
	b, l := model.NewCmdBuff('>', model.CommandBuffer), testListener{}
	b.AddListener(&l)

	b.SetActive(false)
	assert.Equal(t, 0, l.act)
	assert.Equal(t, 1, l.inact)
	assert.False(t, b.IsActive())
}

func TestCmdBuffChanged(t *testing.T) {
	b, l := model.NewCmdBuff('>', model.CommandBuffer), testListener{}
	b.AddListener(&l)

	b.Add('b')
	assert.Equal(t, 0, l.act)
	assert.Equal(t, 0, l.inact)
	assert.Equal(t, "b", l.text)
	assert.Equal(t, "b", b.GetText())

	b.Delete()
	assert.Equal(t, 0, l.act)
	assert.Equal(t, 0, l.inact)
	assert.Equal(t, "", l.text)
	assert.Equal(t, "", b.GetText())

	b.Add('c')
	b.ClearText(true)
	assert.Equal(t, 0, l.act)
	assert.Equal(t, 0, l.inact)
	assert.Equal(t, "", l.text)
	assert.Equal(t, "", b.GetText())

	b.Add('c')
	b.Reset()
	assert.Equal(t, 0, l.act)
	assert.Equal(t, 1, l.inact)
	assert.Equal(t, "", l.text)
	assert.Equal(t, "", b.GetText())
	assert.True(t, b.Empty())
}

func TestCmdBuffAdd(t *testing.T) {
	b := model.NewCmdBuff('>', model.CommandBuffer)

	uu := []struct {
		runes []rune
		cmd   string
	}{
		{[]rune{}, ""},
		{[]rune{'a'}, "a"},
		{[]rune{'a', 'b', 'c'}, "abc"},
	}

	for _, u := range uu {
		for _, r := range u.runes {
			b.Add(r)
		}
		assert.Equal(t, u.cmd, b.GetText())
		b.Reset()
	}
}

func TestCmdBuffDel(t *testing.T) {
	b := model.NewCmdBuff('>', model.CommandBuffer)

	uu := []struct {
		runes []rune
		cmd   string
	}{
		{[]rune{}, ""},
		{[]rune{'a'}, ""},
		{[]rune{'a', 'b', 'c'}, "ab"},
	}

	for _, u := range uu {
		for _, r := range u.runes {
			b.Add(r)
		}
		b.Delete()
		assert.Equal(t, u.cmd, b.GetText())
		b.Reset()
	}
}

func TestCmdBuffEmpty(t *testing.T) {
	b := model.NewCmdBuff('>', model.CommandBuffer)

	uu := []struct {
		runes []rune
		empty bool
	}{
		{[]rune{}, true},
		{[]rune{'a'}, false},
		{[]rune{'a', 'b', 'c'}, false},
	}

	for _, u := range uu {
		for _, r := range u.runes {
			b.Add(r)
		}
		assert.Equal(t, u.empty, b.Empty())
		b.Reset()
	}
}
