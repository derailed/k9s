// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"sort"
)

// SuggestionListener listens for suggestions.
type SuggestionListener interface {
	BuffWatcher

	// SuggestionChanged notifies suggestion changes.
	SuggestionChanged(text, sugg string)
}

// SuggestionFunc produces suggestions.
type SuggestionFunc func(text string) sort.StringSlice

// FishBuff represents a suggestion buffer.
type FishBuff struct {
	*CmdBuff

	suggestionFn    SuggestionFunc
	suggestions     []string
	suggestionIndex int
}

// NewFishBuff returns a new command buffer.
func NewFishBuff(key rune, kind BufferKind) *FishBuff {
	return &FishBuff{
		CmdBuff:         NewCmdBuff(key, kind),
		suggestionIndex: -1,
	}
}

// PrevSuggestion returns the prev suggestion.
func (f *FishBuff) PrevSuggestion() (string, bool) {
	if len(f.suggestions) == 0 {
		return "", false
	}

	if f.suggestionIndex < 0 {
		f.suggestionIndex = 0
	} else {
		f.suggestionIndex--
	}
	if f.suggestionIndex < 0 {
		f.suggestionIndex = len(f.suggestions) - 1
	}

	return f.suggestions[f.suggestionIndex], true
}

// NextSuggestion returns the next suggestion.
func (f *FishBuff) NextSuggestion() (string, bool) {
	if len(f.suggestions) == 0 {
		return "", false
	}

	if f.suggestionIndex < 0 {
		f.suggestionIndex = 0
	} else {
		f.suggestionIndex++
	}
	if f.suggestionIndex >= len(f.suggestions) {
		f.suggestionIndex = 0
	}

	return f.suggestions[f.suggestionIndex], true
}

// ClearSuggestions clear out all suggestions.
func (f *FishBuff) ClearSuggestions() {
	if len(f.suggestions) > 0 {
		f.suggestions = f.suggestions[:0]
	}
	f.suggestionIndex = -1
}

// CurrentSuggestion returns the current suggestion.
func (f *FishBuff) CurrentSuggestion() (string, bool) {
	if len(f.suggestions) == 0 || f.suggestionIndex < 0 || f.suggestionIndex >= len(f.suggestions) {
		return "", false
	}

	return f.suggestions[f.suggestionIndex], true
}

// AutoSuggests returns true if model implements auto suggestions.
func (f *FishBuff) AutoSuggests() bool {
	return true
}

// Suggestions returns suggestions.
func (f *FishBuff) Suggestions() []string {
	if f.suggestionFn != nil {
		return f.suggestionFn(string(f.buff))
	}
	return nil
}

// SetSuggestionFn sets up suggestions.
func (f *FishBuff) SetSuggestionFn(fn SuggestionFunc) {
	f.suggestionFn = fn
}

// Notify publish suggestions to all listeners.
func (f *FishBuff) Notify(delete bool) {
	if f.suggestionFn == nil {
		return
	}
	f.fireSuggestionChanged(f.suggestionFn(string(f.buff)))
}

// Add adds a new character to the buffer.
func (f *FishBuff) Add(r rune) {
	f.CmdBuff.Add(r)
	f.Notify(false)
}

// Delete removes the last character from the buffer.
func (f *FishBuff) Delete() {
	f.CmdBuff.Delete()
	f.Notify(true)
}

func (f *FishBuff) fireSuggestionChanged(ss []string) {
	f.suggestions, f.suggestionIndex = ss, 0

	var suggest string
	if len(ss) == 0 {
		f.suggestionIndex = -1
	} else {
		suggest = ss[f.suggestionIndex]
	}
	f.SetText(f.GetText(), suggest)
}
