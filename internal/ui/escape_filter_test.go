// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"testing"
	"time"

	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
)

func makeRuneEvt(r rune) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, r, tcell.ModNone)
}

func makeAltRuneEvt(r rune) *tcell.EventKey {
	return tcell.NewEventKey(tcell.KeyRune, r, tcell.ModAlt)
}

func TestEscapeFilter_OSC10WithAlt(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	assert.True(t, f.Filter(makeAltRuneEvt(']')), "Alt+] should be filtered (OSC start)")
	for _, r := range "10;rgb:fafa/f9f9/f6f6" {
		assert.True(t, f.Filter(makeRuneEvt(r)), "OSC body char %c should be filtered", r)
	}
	assert.True(t, f.Filter(makeAltRuneEvt('\\')), "Alt+\\ should be filtered (ST)")

	assert.False(t, f.Filter(makeRuneEvt('a')), "normal char after OSC should pass through")
}

func TestEscapeFilter_OSC11WithAlt(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	assert.True(t, f.Filter(makeAltRuneEvt(']')))
	for _, r := range "11;rgb:1212/1212/1212" {
		assert.True(t, f.Filter(makeRuneEvt(r)))
	}
	assert.True(t, f.Filter(makeAltRuneEvt('\\')))

	assert.False(t, f.Filter(makeRuneEvt('z')))
}

func TestEscapeFilter_OSC10BareBracket(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	// When tcell doesn't set ModAlt on ], the ] passes through (false),
	// but after confirming ]10; pattern, the rest is filtered.
	assert.False(t, f.Filter(makeRuneEvt(']')), "bare ] passes through (probing)")
	assert.False(t, f.Filter(makeRuneEvt('1')), "1 passes through (probing)")
	assert.False(t, f.Filter(makeRuneEvt('0')), "0 passes through (probing)")
	// ; confirms it's an OSC color response — filter from here
	assert.True(t, f.Filter(makeRuneEvt(';')), "; confirms OSC, should be filtered")
	for _, r := range "rgb:fafa/f9f9/f6f6" {
		assert.True(t, f.Filter(makeRuneEvt(r)), "OSC body char %c should be filtered", r)
	}
	assert.True(t, f.Filter(makeAltRuneEvt('\\')), "Alt+\\ terminator should be filtered")

	assert.False(t, f.Filter(makeRuneEvt('a')))
}

func TestEscapeFilter_OSC11BareFollowedByCPR(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	// ]11;rgb:2828/2c2c/3434 followed by bare [ CSI
	assert.False(t, f.Filter(makeRuneEvt(']')))
	assert.False(t, f.Filter(makeRuneEvt('1')))
	assert.False(t, f.Filter(makeRuneEvt('1')))
	assert.True(t, f.Filter(makeRuneEvt(';')), "; confirms OSC")
	for _, r := range "rgb:2828/2c2c/3434" {
		assert.True(t, f.Filter(makeRuneEvt(r)))
	}
	// Bare [ in OSC state transitions directly to CSI — filtered
	assert.True(t, f.Filter(makeRuneEvt('[')), "bare [ in OSC state enters CSI")
	for _, r := range "6;14" {
		assert.True(t, f.Filter(makeRuneEvt(r)))
	}
	assert.True(t, f.Filter(makeRuneEvt('R')))
	assert.False(t, f.Filter(makeRuneEvt('x')))
}

func TestEscapeFilter_BareOSCThenBareCPR(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	// Simulate: ]11;rgb:2828/2c2c/3434 then bare [6;14R
	// The ] starts probing, ;confirms OSC, then bare [ transitions to CSI
	assert.False(t, f.Filter(makeRuneEvt(']')))
	assert.False(t, f.Filter(makeRuneEvt('1')))
	assert.False(t, f.Filter(makeRuneEvt('1')))
	assert.True(t, f.Filter(makeRuneEvt(';')))
	for _, r := range "rgb:2828/2c2c/3434" {
		assert.True(t, f.Filter(makeRuneEvt(r)))
	}
	// Bare [ after OSC body — should transition to CSI
	assert.True(t, f.Filter(makeRuneEvt('[')))
	for _, r := range "6;14" {
		assert.True(t, f.Filter(makeRuneEvt(r)))
	}
	assert.True(t, f.Filter(makeRuneEvt('R')))

	assert.False(t, f.Filter(makeRuneEvt('x')))
}

func TestEscapeFilter_CPRWithAlt(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	assert.True(t, f.Filter(makeAltRuneEvt('[')))
	for _, r := range "12;149" {
		assert.True(t, f.Filter(makeRuneEvt(r)))
	}
	assert.True(t, f.Filter(makeRuneEvt('R')))

	assert.False(t, f.Filter(makeRuneEvt('x')))
}

func TestEscapeFilter_FullBurstMixed(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	// Real-world: OSC10 (bare]) + CPR + OSC11 (Alt+]) + CPR
	// First OSC10 without Alt — ]10;rgb:ffff/ffff/ffff
	assert.False(t, f.Filter(makeRuneEvt(']')))
	assert.False(t, f.Filter(makeRuneEvt('1')))
	assert.False(t, f.Filter(makeRuneEvt('0')))
	assert.True(t, f.Filter(makeRuneEvt(';'))) // confirmed
	for _, r := range "rgb:ffff/ffff/ffff" {
		assert.True(t, f.Filter(makeRuneEvt(r)))
	}

	// CPR with bare [ (from within OSC state)
	assert.True(t, f.Filter(makeRuneEvt('[')))
	for _, r := range "6;15" {
		assert.True(t, f.Filter(makeRuneEvt(r)))
	}
	assert.True(t, f.Filter(makeRuneEvt('R')))

	// Second OSC11 with Alt+]
	assert.True(t, f.Filter(makeAltRuneEvt(']')))
	for _, r := range "11;rgb:2828/2c2c/3434" {
		assert.True(t, f.Filter(makeRuneEvt(r)))
	}

	// CPR with Alt+[
	assert.True(t, f.Filter(makeAltRuneEvt('[')))
	for _, r := range "6;14" {
		assert.True(t, f.Filter(makeRuneEvt(r)))
	}
	assert.True(t, f.Filter(makeRuneEvt('R')))

	// Normal input after
	assert.False(t, f.Filter(makeRuneEvt('h')))
	assert.False(t, f.Filter(makeRuneEvt('i')))
}

func TestEscapeFilter_OSCWithBELTerminator(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	assert.True(t, f.Filter(makeAltRuneEvt(']')))
	for _, r := range "10;rgb:fafa/f9f9/f6f6" {
		assert.True(t, f.Filter(makeRuneEvt(r)))
	}
	assert.True(t, f.Filter(makeRuneEvt(0x07)), "BEL should terminate OSC")

	assert.False(t, f.Filter(makeRuneEvt('a')))
}

func TestEscapeFilter_NormalTypingUnaffected(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	normalChars := "hello world/pod:123[testR;foo"
	for _, r := range normalChars {
		time.Sleep(10 * time.Millisecond) // simulate typing speed
		assert.False(t, f.Filter(makeRuneEvt(r)),
			"normal char %c should not be filtered", r)
	}
}

func TestEscapeFilter_BareRightBracketThenNormal(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	// User types ] then something that's NOT 1 — should pass through
	assert.False(t, f.Filter(makeRuneEvt(']')))
	assert.False(t, f.Filter(makeRuneEvt('a')), "non-1 after ] should pass through")

	// User types ]1 then something that's NOT 0 or 1 — should pass through
	f2 := NewEscapeSequenceFilter(nil)
	assert.False(t, f2.Filter(makeRuneEvt(']')))
	assert.False(t, f2.Filter(makeRuneEvt('1')))
	assert.False(t, f2.Filter(makeRuneEvt('x')), "non-0/1 after ]1 should pass through")

	// User types ]10 then something that's NOT ; — should pass through
	f3 := NewEscapeSequenceFilter(nil)
	assert.False(t, f3.Filter(makeRuneEvt(']')))
	assert.False(t, f3.Filter(makeRuneEvt('1')))
	assert.False(t, f3.Filter(makeRuneEvt('0')))
	assert.False(t, f3.Filter(makeRuneEvt('x')), "non-; after ]10 should pass through")
}

func TestEscapeFilter_ProbeTimeoutResets(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	// Start probing with bare ]
	assert.False(t, f.Filter(makeRuneEvt(']')))

	// Wait longer than probe timeout
	time.Sleep(escProbeTimeout + 20*time.Millisecond)

	// State should reset — normal chars pass through
	assert.False(t, f.Filter(makeRuneEvt('1')))
}

func TestEscapeFilter_BurstTimeoutResets(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	assert.True(t, f.Filter(makeAltRuneEvt(']')))
	assert.True(t, f.Filter(makeRuneEvt('1')))

	time.Sleep(escBurstTimeout + 20*time.Millisecond)

	assert.False(t, f.Filter(makeRuneEvt('h')))
}

func TestEscapeFilter_MalformedCSIResets(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	assert.True(t, f.Filter(makeAltRuneEvt('[')))
	assert.True(t, f.Filter(makeRuneEvt('1')))

	assert.False(t, f.Filter(makeRuneEvt('å')),
		"non-ASCII char should reset CSI filter")

	assert.False(t, f.Filter(makeRuneEvt('a')))
}

// TestEscapeFilter_GhosttyBypass simulates the case where Ghostty consumes
// the OSC prefix (\x1b]10;rgb) and passes only :ffff/ffff/ffff to the app.
// The : arrives first (with no prior context), followed by hex chars in a
// rapid burst.
func TestEscapeFilter_GhosttyBypass(t *testing.T) {
	undoCalled := false
	f := NewEscapeSequenceFilter(func() { undoCalled = true })

	// : arrives — enters colon probe, passes through
	assert.False(t, f.Filter(makeRuneEvt(':')))

	// Hex digit arrives within burst gap — confirms escape residue
	time.Sleep(1 * time.Millisecond) // within escBurstGap (5ms)
	assert.True(t, f.Filter(makeRuneEvt('f')), "hex after : in burst should be filtered")
	assert.True(t, undoCalled, "undo should have been called to deactivate command mode")

	// Rest of the RGB body is filtered
	for _, r := range "fff/ffff/ffff" {
		assert.True(t, f.Filter(makeRuneEvt(r)), "body char %c should be filtered", r)
	}

	// Normal typing after the burst ends (with a gap)
	time.Sleep(10 * time.Millisecond)
	assert.False(t, f.Filter(makeRuneEvt('a')))
}

func TestEscapeFilter_CSIVariousFinalBytes(t *testing.T) {
	f := NewEscapeSequenceFilter(nil)

	for _, final := range []rune{'R', 'n', 'c', 't'} {
		assert.True(t, f.Filter(makeAltRuneEvt('[')))
		assert.True(t, f.Filter(makeRuneEvt('1')))
		assert.True(t, f.Filter(makeRuneEvt(';')))
		assert.True(t, f.Filter(makeRuneEvt('2')))
		assert.True(t, f.Filter(makeRuneEvt(final)))
		assert.False(t, f.Filter(makeRuneEvt('x')))
	}
}
