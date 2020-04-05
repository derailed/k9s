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
	suggestion      string
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
func (c *FishBuff) PrevSuggestion() (string, bool) {
	if c.suggestionIndex < 0 {
		return "", false
	}
	c.suggestionIndex--
	if c.suggestionIndex < 0 {
		c.suggestionIndex = len(c.suggestions) - 1
	}
	return c.suggestions[c.suggestionIndex], true
}

// NextSuggestion returns the next suggestion.
func (c *FishBuff) NextSuggestion() (string, bool) {
	if c.suggestionIndex < 0 {
		return "", false
	}
	c.suggestionIndex++
	if c.suggestionIndex >= len(c.suggestions) {
		c.suggestionIndex = 0
	}
	return c.suggestions[c.suggestionIndex], true
}

// ClearSuggestions clear out all suggestions.
func (c *FishBuff) ClearSuggestions() {
	c.suggestion, c.suggestionIndex = "", -1
}

// CurrentSuggestion returns the current suggestion.
func (c *FishBuff) CurrentSuggestion() (string, bool) {
	if c.suggestionIndex < 0 {
		return "", false
	}
	return c.suggestions[c.suggestionIndex], true
}

// AutoSuggests returns true if model implements auto suggestions.
func (c *FishBuff) AutoSuggests() bool {
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
func (f *FishBuff) Notify() {
	if f.suggestionFn == nil {
		return
	}
	cc := f.suggestionFn(string(f.buff))
	f.fireSuggestionChanged(cc)
}

// Add adds a new charater to the buffer.
func (f *FishBuff) Add(r rune) {
	f.CmdBuff.Add(r)
	if f.suggestionFn == nil {
		return
	}
	cc := f.suggestionFn(string(f.buff))
	f.fireSuggestionChanged(cc)
}

// Delete removes the last character from the buffer.
func (f *FishBuff) Delete() {
	f.CmdBuff.Delete()
	if f.suggestionFn == nil {
		return
	}
	cc := f.suggestionFn(string(f.buff))
	f.fireSuggestionChanged(cc)
}

func (f *FishBuff) fireSuggestionChanged(ss []string) {
	f.suggestions, f.suggestionIndex = ss, 0
	if ss == nil {
		f.suggestionIndex, f.suggestion = -1, ""
		return
	}
	f.suggestion = ss[f.suggestionIndex]

	for _, l := range f.listeners {
		if listener, ok := l.(SuggestionListener); ok {
			listener.SuggestionChanged(f.GetText(), f.suggestion)
		}
	}
}
