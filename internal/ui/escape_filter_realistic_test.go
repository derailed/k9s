// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"testing"
	"time"

	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
)

// TestEscapeFilter_RealisticTcellBehavior simulates the exact event sequence
// that derailed/tcell v2.3.1-rc.4 produces when it receives OSC 10/11 color
// responses and CPR cursor position reports.
//
// Key finding from production debug logs (2026-04-03T20:16:41):
// - The FIRST OSC sequence's ] arrives WITHOUT ModAlt (tcell's input buffering
//   causes ESC to be consumed without setting the escaped flag)
// - The SECOND OSC sequence's ] arrives WITH ModAlt (normal behavior)
// - CPR [ characters can arrive either way
//
// The raw terminal response is:
//
//	\x1b]10;rgb:ffff/ffff/ffff\x1b\\\x1b[6;15R\x1b]11;rgb:2828/2c2c/3434\x1b\\\x1b[6;14R
//
// This test constructs the events exactly as observed in production.
func TestEscapeFilter_RealisticTcellBehavior(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	type event struct {
		r      rune
		mod    tcell.ModMask
		expect bool   // true = should be filtered
		desc   string // description for failure message
	}

	// Build the exact event sequence observed in production.
	// First OSC (10) — bare ] without Alt (tcell bug)
	events := []event{
		{']', tcell.ModNone, false, "bare ] (first OSC, no Alt) — passes through"},
		{'1', tcell.ModNone, false, "probing: 1 after ]"},
		{'0', tcell.ModNone, false, "probing: 0 after ]1"},
		{';', tcell.ModNone, true, "; confirms ]10 is OSC — filter starts"},
	}
	// Add the OSC 10 body: rgb:ffff/ffff/ffff
	for _, r := range "rgb:ffff/ffff/ffff" {
		events = append(events, event{r, tcell.ModNone, true, "OSC 10 body"})
	}

	// First CPR — bare [ (ESC consumed without escaped flag, same issue)
	events = append(events, event{'[', tcell.ModNone, true, "bare [ entering CSI from OSC"})
	for _, r := range "6;15" {
		events = append(events, event{r, tcell.ModNone, true, "CPR param"})
	}
	events = append(events, event{'R', tcell.ModNone, true, "CPR final byte"})

	// Second OSC (11) — Alt+] (tcell sets ModAlt correctly this time)
	events = append(events, event{']', tcell.ModAlt, true, "Alt+] (second OSC, with Alt)"})
	for _, r := range "11;rgb:2828/2c2c/3434" {
		events = append(events, event{r, tcell.ModNone, true, "OSC 11 body"})
	}

	// ST terminator — Alt+backslash
	events = append(events, event{'\\', tcell.ModAlt, true, "Alt+\\ (ST terminator)"})

	// Second CPR — Alt+[
	events = append(events, event{'[', tcell.ModAlt, true, "Alt+[ (second CPR)"})
	for _, r := range "6;14" {
		events = append(events, event{r, tcell.ModNone, true, "CPR param"})
	}
	events = append(events, event{'R', tcell.ModNone, true, "CPR final byte"})

	// Run all events through the filter
	var leaked []rune
	for i, e := range events {
		evt := tcell.NewEventKey(tcell.KeyRune, e.r, e.mod)
		got := f.Filter(evt)
		assert.Equal(t, e.expect, got,
			"event %d: rune=%q mod=%d — %s", i, e.r, e.mod, e.desc)
		if !got {
			leaked = append(leaked, e.r)
		}
	}

	// The only chars that should leak are ]10 (the probe prefix)
	assert.Equal(t, []rune{']', '1', '0'}, leaked,
		"only the ]10 probe prefix should leak through")

	// After the sequence, normal input must work
	assert.False(t, f.Filter(tcell.NewEventKey(tcell.KeyRune, 'p', tcell.ModNone)))
	assert.False(t, f.Filter(tcell.NewEventKey(tcell.KeyRune, 'o', tcell.ModNone)))
	assert.False(t, f.Filter(tcell.NewEventKey(tcell.KeyRune, 'd', tcell.ModNone)))
}

// TestEscapeFilter_RealisticAllBare simulates the worst case where ALL
// escape sequences arrive without ModAlt on ] and [.
func TestEscapeFilter_RealisticAllBare(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	type event struct {
		r      rune
		mod    tcell.ModMask
		expect bool
	}

	events := []event{
		// OSC 10 — bare ]
		{']', tcell.ModNone, false},
		{'1', tcell.ModNone, false},
		{'0', tcell.ModNone, false},
		{';', tcell.ModNone, true}, // confirmed
	}
	for _, r := range "rgb:ffff/ffff/ffff" {
		events = append(events, event{r, tcell.ModNone, true})
	}
	// CPR — bare [
	events = append(events, event{'[', tcell.ModNone, true}) // from OSC state
	for _, r := range "6;15" {
		events = append(events, event{r, tcell.ModNone, true})
	}
	events = append(events, event{'R', tcell.ModNone, true})

	// OSC 11 — bare ] after CSI completed (back in escNone)
	events = append(events, event{']', tcell.ModNone, false})
	events = append(events, event{'1', tcell.ModNone, false})
	events = append(events, event{'1', tcell.ModNone, false})
	events = append(events, event{';', tcell.ModNone, true}) // confirmed
	for _, r := range "rgb:2828/2c2c/3434" {
		events = append(events, event{r, tcell.ModNone, true})
	}
	// CPR — bare [
	events = append(events, event{'[', tcell.ModNone, true})
	for _, r := range "6;14" {
		events = append(events, event{r, tcell.ModNone, true})
	}
	events = append(events, event{'R', tcell.ModNone, true})

	var leaked []rune
	for _, e := range events {
		evt := tcell.NewEventKey(tcell.KeyRune, e.r, e.mod)
		got := f.Filter(evt)
		assert.Equal(t, e.expect, got, "rune=%q mod=%d", e.r, e.mod)
		if !got {
			leaked = append(leaked, e.r)
		}
	}

	// Both ]10 and ]11 probe prefixes leak
	assert.Equal(t, []rune{']', '1', '0', ']', '1', '1'}, leaked,
		"only probe prefixes should leak")

	// Verify the leaked chars don't contain / or : (which activate filter/command mode)
	for _, r := range leaked {
		assert.NotEqual(t, '/', r, "/ must never leak (activates filter mode)")
		assert.NotEqual(t, ':', r, ": must never leak (activates command mode)")
	}
}

// TestEscapeFilter_RealisticNoActivation verifies that the critical characters
// / (filter mode) and : (command mode) are NEVER in the leaked set, regardless
// of how tcell delivers the events.
func TestEscapeFilter_RealisticNoActivation(t *testing.T) {
	// Test both scenarios: mixed Alt/bare and all-bare
	for _, scenario := range []string{"mixed", "all-bare"} {
		t.Run(scenario, func(t *testing.T) {
			f := NewEscapeSequenceFilter(nil)

			// Full terminal response bytes (what the terminal actually sends)
			// \x1b]10;rgb:ffff/ffff/ffff\x1b\\\x1b[6;15R\x1b]11;rgb:2828/2c2c/3434\x1b\\\x1b[6;14R
			//
			// After tcell processing, we get KeyRune events.
			// The : in rgb: and / in ffff/ffff are the dangerous chars.

			var events []simEvent

			if scenario == "mixed" {
				events = buildMixedEvents()
			} else {
				events = buildAllBareEvents()
			}

			var leaked []rune
			for _, e := range events {
				evt := tcell.NewEventKey(tcell.KeyRune, e.r, e.mod)
				if !f.Filter(evt) {
					leaked = append(leaked, e.r)
				}
			}

			// Critical assertion: no activation characters leak
			for _, r := range leaked {
				assert.NotEqual(t, '/', r,
					"/ must not leak — it activates filter mode in k9s")
				assert.NotEqual(t, ':', r,
					": must not leak — it activates command mode in k9s")
				assert.NotEqual(t, ';', r,
					"; should not leak — it's part of the OSC body")
			}
		})
	}
}

// TestEscapeFilter_RealisticWithDelay verifies the filter handles events
// with realistic timing gaps between them (simulating PostEventWait blocking).
func TestEscapeFilter_RealisticWithDelay(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping timing-sensitive test in short mode")
	}

	f := NewEscapeSequenceFilter(nil)

	// Send bare ]10; with small gaps (simulating busy event loop)
	assert.False(t, f.Filter(makeRuneEvt(']')))
	time.Sleep(5 * time.Millisecond)
	assert.False(t, f.Filter(makeRuneEvt('1')))
	time.Sleep(5 * time.Millisecond)
	assert.False(t, f.Filter(makeRuneEvt('0')))
	time.Sleep(5 * time.Millisecond)
	assert.True(t, f.Filter(makeRuneEvt(';')), "; should confirm OSC")

	// Body with gaps
	for _, r := range "rgb:" {
		time.Sleep(2 * time.Millisecond)
		assert.True(t, f.Filter(makeRuneEvt(r)), "char %c in body should be filtered", r)
	}

	// Simulate a longer pause (draw cycle) mid-body
	time.Sleep(80 * time.Millisecond)
	for _, r := range "ffff/ffff/ffff" {
		assert.True(t, f.Filter(makeRuneEvt(r)), "char %c after pause should still be filtered", r)
	}
}

// helpers

type simEvent struct {
	r   rune
	mod tcell.ModMask
}

func buildMixedEvents() []simEvent {
	var evts []simEvent

	// OSC 10 — bare ]
	evts = append(evts, simEvent{']', tcell.ModNone})
	for _, r := range "10;rgb:ffff/ffff/ffff" {
		evts = append(evts, simEvent{r, tcell.ModNone})
	}
	// CPR — bare [
	evts = append(evts, simEvent{'[', tcell.ModNone})
	for _, r := range "6;15" {
		evts = append(evts, simEvent{r, tcell.ModNone})
	}
	evts = append(evts, simEvent{'R', tcell.ModNone})

	// OSC 11 — Alt+]
	evts = append(evts, simEvent{']', tcell.ModAlt})
	for _, r := range "11;rgb:2828/2c2c/3434" {
		evts = append(evts, simEvent{r, tcell.ModNone})
	}
	evts = append(evts, simEvent{'\\', tcell.ModAlt})

	// CPR — Alt+[
	evts = append(evts, simEvent{'[', tcell.ModAlt})
	for _, r := range "6;14" {
		evts = append(evts, simEvent{r, tcell.ModNone})
	}
	evts = append(evts, simEvent{'R', tcell.ModNone})

	return evts
}

func buildAllBareEvents() []simEvent {
	var evts []simEvent

	// OSC 10 — bare ]
	evts = append(evts, simEvent{']', tcell.ModNone})
	for _, r := range "10;rgb:ffff/ffff/ffff" {
		evts = append(evts, simEvent{r, tcell.ModNone})
	}
	evts = append(evts, simEvent{'[', tcell.ModNone})
	for _, r := range "6;15" {
		evts = append(evts, simEvent{r, tcell.ModNone})
	}
	evts = append(evts, simEvent{'R', tcell.ModNone})

	// OSC 11 — bare ]
	evts = append(evts, simEvent{']', tcell.ModNone})
	for _, r := range "11;rgb:2828/2c2c/3434" {
		evts = append(evts, simEvent{r, tcell.ModNone})
	}
	evts = append(evts, simEvent{'[', tcell.ModNone})
	for _, r := range "6;14" {
		evts = append(evts, simEvent{r, tcell.ModNone})
	}
	evts = append(evts, simEvent{'R', tcell.ModNone})

	return evts
}
