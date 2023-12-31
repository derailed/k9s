// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/sahilm/fuzzy"
	"github.com/stretchr/testify/assert"
)

func TestNewText(t *testing.T) {
	m := model.NewText()

	lis := textLis{}
	m.AddListener(&lis)

	m.SetText("Hello World\nBumbleBeeTuna")

	assert.Equal(t, 1, lis.changed)
	assert.Equal(t, 2, lis.lines)
	assert.Equal(t, 0, lis.filtered)
	assert.Equal(t, 0, lis.matches)
}

func TestTextFilterRXMatch(t *testing.T) {
	m := model.NewText()

	lis := textLis{}
	m.AddListener(&lis)

	m.SetText("Hello World\nBumbleBeeTuna")
	m.Filter("world")

	assert.Equal(t, 1, lis.changed)
	assert.Equal(t, 2, lis.lines)
	assert.Equal(t, 1, lis.filtered)
	assert.Equal(t, 1, lis.matches)
	assert.Equal(t, 6, lis.index)
}

func TestTextFilterFuzzyMatch(t *testing.T) {
	m := model.NewText()

	lis := textLis{}
	m.AddListener(&lis)

	m.SetText("Hello World\nBumbleBeeTuna")
	m.Filter("-f world")

	assert.Equal(t, 1, lis.changed)
	assert.Equal(t, 2, lis.lines)
	assert.Equal(t, 1, lis.filtered)
	assert.Equal(t, 1, lis.matches)
	assert.Equal(t, 6, lis.index)
}

func TestTextFilterNoMatch(t *testing.T) {
	m := model.NewText()

	lis := textLis{}
	m.AddListener(&lis)

	m.SetText("Hello World\nBumbleBeeTuna")
	m.Filter("blee")

	assert.Equal(t, 1, lis.changed)
	assert.Equal(t, 2, lis.lines)
	assert.Equal(t, 1, lis.filtered)
	assert.Equal(t, 0, lis.matches)
	assert.Equal(t, 0, lis.index)
}

// Helpers...

type textLis struct {
	changed, filtered, matches, lines, index int
}

func (l *textLis) TextChanged(ll []string) {
	l.lines = len(ll)
	l.changed++
}

func (l *textLis) TextFiltered(ll []string, mm fuzzy.Matches) {
	l.matches = len(mm)
	l.filtered++
	if len(mm) > 0 && len(mm[0].MatchedIndexes) > 0 {
		l.index = mm[0].MatchedIndexes[0]
	}
}
