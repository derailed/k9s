// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"time"

	"github.com/derailed/tcell/v2"
)

// escFilterState tracks the state of escape sequence detection.
type escFilterState int

const (
	escNone      escFilterState = iota
	escInOSC                    // confirmed OSC sequence, accumulating body
	escInCSI                    // confirmed CSI sequence, accumulating body
	escSawBrack                 // saw bare ] (no Alt), might be OSC start
	escSawBrack1                // saw ] then 1 — waiting for 0 or 1
	escSawBrackN                // saw ]10 or ]11 — waiting for ;
	escInBurst                  // in a rapid burst of escape-sequence-like chars
	escSawColon                 // saw : or / rapidly, might be OSC body start
)

const (
	// escBurstTimeout defines how long to wait before resetting filter state.
	escBurstTimeout = 2 * time.Second

	// escProbeTimeout is a short timeout for speculative detection states.
	escProbeTimeout = 100 * time.Millisecond

	// escBurstGap is the maximum time between consecutive events to be
	// considered part of a terminal response burst. Terminal responses
	// deliver all characters within ~1ms. Real typing is 30ms+ apart.
	escBurstGap = 5 * time.Millisecond
)

// EscapeUndoFunc is called when the filter retroactively determines
// that a character that already passed through (like ':') was part of
// an escape sequence. The caller should deactivate any mode that was
// activated and clear the buffer.
type EscapeUndoFunc func()

// EscapeSequenceFilter detects and discards terminal escape sequence
// responses (OSC 10/11 color queries, CPR cursor position reports)
// that leak through the tcell input parser as individual KeyRune events.
//
// Detection uses three strategies:
//  1. Alt+] / Alt+[ signals (when tcell correctly sets ModAlt)
//  2. Pattern matching: bare ] followed by 10; or 11;
//  3. Timing-based burst detection: when : or / is followed by hex digits
//     within 5ms, it's escape sequence residue from a terminal that
//     partially consumed the OSC prefix.
type EscapeSequenceFilter struct {
	state       escFilterState
	lastEvent   time.Time
	pendNum     rune
	burstCount  int
	burstEscLen int
	undoFn      EscapeUndoFunc
}

// NewEscapeSequenceFilter creates a new filter. The undoFn is called
// when the filter retroactively determines that a : or / that already
// passed through was part of an escape sequence.
func NewEscapeSequenceFilter(undoFn EscapeUndoFunc) *EscapeSequenceFilter {
	return &EscapeSequenceFilter{undoFn: undoFn}
}

// IsActive returns true if the filter is currently tracking a sequence.
func (f *EscapeSequenceFilter) IsActive() bool {
	return f.state != escNone
}

// isEscapeResidue returns true if the rune could be part of a terminal
// escape sequence response (OSC color or CPR).
func isEscapeResidue(r rune) bool {
	if r >= '0' && r <= '9' || r >= 'a' && r <= 'f' || r >= 'A' && r <= 'F' {
		return true
	}
	switch r {
	case '/', ':', ';', '[', ']', '\\', 'R', 'r', 'g', 'b':
		return true
	}
	return false
}

// Filter examines a tcell.EventKey and returns true if the event
// should be discarded (it is part of a terminal response sequence).
// Only call this for KeyRune events.
func (f *EscapeSequenceFilter) Filter(evt *tcell.EventKey) bool {
	now := time.Now()
	elapsed := now.Sub(f.lastEvent)

	timeout := escBurstTimeout
	if f.state >= escSawBrack && f.state <= escSawBrackN || f.state == escSawColon {
		timeout = escProbeTimeout
	}

	if f.state != escNone && elapsed > timeout {
		f.reset()
	}

	if elapsed < escBurstGap && f.lastEvent.Unix() > 0 {
		f.burstCount++
	} else {
		f.burstCount = 1
		f.burstEscLen = 0
	}
	f.lastEvent = now

	r := evt.Rune()
	mod := evt.Modifiers()

	if isEscapeResidue(r) {
		f.burstEscLen++
	}

	switch f.state {
	case escNone:
		if mod&tcell.ModAlt != 0 && r == ']' {
			f.state = escInOSC
			return true
		}
		if mod&tcell.ModAlt != 0 && r == '[' {
			f.state = escInCSI
			return true
		}
		if mod == tcell.ModNone && r == ']' {
			f.state = escSawBrack
			return false
		}
		if mod == tcell.ModNone && (r == ':' || r == '/') {
			f.state = escSawColon
			return false
		}
		return false

	case escSawBrack:
		if r == '1' && mod == tcell.ModNone {
			f.pendNum = r
			f.state = escSawBrack1
			return false
		}
		f.reset()
		return false

	case escSawBrack1:
		if (r == '0' || r == '1') && mod == tcell.ModNone {
			f.pendNum = r
			f.state = escSawBrackN
			return false
		}
		f.reset()
		return false

	case escSawBrackN:
		if r == ';' && mod == tcell.ModNone {
			f.state = escInOSC
			return true
		}
		f.reset()
		return false

	case escInOSC:
		if mod&tcell.ModAlt != 0 && r == '\\' {
			f.reset()
			return true
		}
		if r == 0x07 {
			f.reset()
			return true
		}
		if mod&tcell.ModAlt != 0 && r == ']' {
			return true
		}
		if r == '[' {
			f.state = escInCSI
			return true
		}
		return true

	case escInCSI:
		if r >= '0' && r <= '?' {
			return true
		}
		if r >= ' ' && r <= '/' {
			return true
		}
		if r >= '@' && r <= '~' {
			f.reset()
			return true
		}
		if mod&tcell.ModAlt != 0 && r == ']' {
			f.state = escInOSC
			return true
		}
		if mod&tcell.ModAlt != 0 && r == '[' {
			return true
		}
		if mod == tcell.ModNone && r == ']' {
			f.state = escSawBrack
			return false
		}
		f.reset()
		return false

	case escSawColon:
		isHex := (r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
		if isHex && elapsed < escBurstGap && elapsed > 0 {
			if f.undoFn != nil {
				f.undoFn()
			}
			f.state = escInBurst
			return true
		}
		f.reset()
		return false

	case escInBurst:
		if elapsed < escBurstGap && isEscapeResidue(r) {
			return true
		}
		if mod&tcell.ModAlt != 0 && r == '\\' {
			f.reset()
			return true
		}
		if mod&tcell.ModAlt != 0 && r == '[' {
			f.state = escInCSI
			return true
		}
		if r == '[' && elapsed < escBurstGap {
			f.state = escInCSI
			return true
		}
		f.reset()
		return false
	}

	return false
}

func (f *EscapeSequenceFilter) reset() {
	f.state = escNone
	f.pendNum = 0
}
