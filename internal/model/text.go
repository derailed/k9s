// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/sahilm/fuzzy"
)

// Filterable represents an entity that can be filtered.
type Filterable interface {
	Filter(string)
	ClearFilter()
}

// Textable represents a text resource.
type Textable interface {
	Peek() []string
	SetText(string)
	AddListener(TextListener)
	RemoveListener(TextListener)
}

// TextListener represents a text model listener.
type TextListener interface {
	// TextChanged notifies the model changed.
	TextChanged([]string)

	// TextFiltered notifies when the filter changed.
	TextFiltered([]string, fuzzy.Matches)
}

// Text represents a text model.
type Text struct {
	lines     []string
	listeners []TextListener
	query     string
}

// NewText returns a new model.
func NewText() *Text {
	return &Text{}
}

// Peek returns the current model state.
func (t *Text) Peek() []string {
	return t.lines
}

// ClearFilter clear out filter.
func (t *Text) ClearFilter() {
	t.query = ""
	t.filterChanged(t.lines)
}

// Filter filters out the text.
func (t *Text) Filter(q string) {
	t.query = q
	t.filterChanged(t.lines)
}

// SetText sets the current model content.
func (t *Text) SetText(buff string) {
	t.lines = strings.Split(buff, "\n")
	t.fireTextChanged(t.lines)
}

// AddListener adds a new model listener.
func (t *Text) AddListener(listener TextListener) {
	t.listeners = append(t.listeners, listener)
}

// RemoveListener delete a listener from the list.
func (t *Text) RemoveListener(listener TextListener) {
	victim := -1
	for i, lis := range t.listeners {
		if lis == listener {
			victim = i
			break
		}
	}

	if victim >= 0 {
		t.listeners = append(t.listeners[:victim], t.listeners[victim+1:]...)
	}
}

func (t *Text) filterChanged(lines []string) {
	t.fireTextFiltered(lines, t.filter(t.query, lines))
}

func (t *Text) fireTextChanged(lines []string) {
	for _, lis := range t.listeners {
		lis.TextChanged(lines)
	}
}

func (t *Text) fireTextFiltered(lines []string, matches fuzzy.Matches) {
	for _, lis := range t.listeners {
		lis.TextFiltered(lines, matches)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func (t *Text) filter(q string, lines []string) fuzzy.Matches {
	if q == "" {
		return nil
	}
	if f, ok := internal.IsFuzzySelector(q); ok {
		return t.fuzzyFilter(strings.TrimSpace(f), lines)
	}
	return rxFilter(q, lines)
}

func (*Text) fuzzyFilter(q string, lines []string) fuzzy.Matches {
	return fuzzy.Find(q, lines)
}
