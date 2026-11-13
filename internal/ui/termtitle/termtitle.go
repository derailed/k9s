// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

// Package termtitle sets the title of the terminal k9s runs in and puts the
// original title back on exit.
package termtitle

import (
	"strings"
	"sync"
	"unicode"
)

// Title manages the terminal title. It is safe for concurrent use.
type Title struct {
	mx      sync.Mutex
	current string
	saved   bool
	backend backend
}

// backend writes titles to the terminal. Each OS provides its own.
type backend interface {
	// save remembers the current terminal title so restore can put it back.
	save() error
	// set changes the terminal title.
	set(title string) error
	// restore puts back the title remembered by save.
	restore() error
}

// New returns a title manager for the current terminal.
func New() *Title {
	return &Title{backend: newBackend()}
}

// Set changes the terminal title. The original title is saved on first use.
func (t *Title) Set(title string) {
	title = Sanitize(title)

	t.mx.Lock()
	defer t.mx.Unlock()

	if title == t.current {
		return
	}
	if !t.saved {
		if err := t.backend.save(); err != nil {
			return
		}
		t.saved = true
	}
	if err := t.backend.set(title); err != nil {
		return
	}
	t.current = title
}

// Restore puts back the title the terminal had before the first Set.
// It does nothing if Set was never called.
func (t *Title) Restore() {
	t.mx.Lock()
	defer t.mx.Unlock()

	if !t.saved {
		return
	}
	_ = t.backend.restore()
	t.saved, t.current = false, ""
}

// Sanitize removes control characters so a title cannot inject escape sequences.
func Sanitize(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
}
