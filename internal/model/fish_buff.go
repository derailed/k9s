package model

import (
	"sort"
)

// SuggestionListener listens for suggestions.
type SuggestionListener interface {
	BuffWatcher

	// SuggestionChanged notifies suggestion changes.
	SuggestionChanged([]string)
}

// SuggestionFunc produces suggestions.
type SuggestionFunc func(s string) sort.StringSlice

// FishBuff represents a suggestion buffer.
type FishBuff struct {
	*CmdBuff

	suggestionFn SuggestionFunc
}

// NewFishBuff returns a new command buffer.
func NewFishBuff(key rune, kind BufferKind) *FishBuff {
	return &FishBuff{CmdBuff: NewCmdBuff(key, kind)}
}

// SetSuggestionFn sets up suggestions.
func (f *FishBuff) SetSuggestionFn(fn SuggestionFunc) {
	f.suggestionFn = fn
}

// Delete removes the last character from the buffer.
func (f *FishBuff) Delete() {
	f.CmdBuff.Delete()
	if f.suggestionFn == nil {
		return
	}
	cc := f.suggestionFn(string(f.buff))
	f.fireSuggest(cc)
}

// Add adds a new charater to the buffer.
func (f *FishBuff) Add(r rune) {
	f.CmdBuff.Add(r)
	if f.suggestionFn == nil {
		return
	}
	cc := f.suggestionFn(string(f.buff))
	f.fireSuggest(cc)
}

func (f *FishBuff) fireSuggest(cc []string) {
	for _, l := range f.listeners {
		if s, ok := l.(SuggestionListener); ok {
			s.SuggestionChanged(cc)
		}
	}
}
