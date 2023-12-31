// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package vul

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_scorerAdd(t *testing.T) {
	uu := map[string]struct {
		b, b1, e scorer
	}{
		"zero": {},
		"same": {
			b:  scorer(0x80),
			b1: scorer(0x80),
			e:  scorer(0x80),
		},
		"c+h": {
			b:  scorer(0x80),
			b1: scorer(0x40),
			e:  scorer(0xC0),
		},
		"ch+hm": {
			b:  scorer(0xc0),
			b1: scorer(0xa0),
			e:  scorer(0xe0),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.b.Add(u.b1))
		})
	}
}

func Test_scorerFromTally(t *testing.T) {
	uu := map[string]struct {
		tt tally
		b  scorer
	}{
		"zero": {},
		"critical": {
			tt: tally{29, 0, 0, 0, 0, 0, 0},
			b:  scorer(0x80),
		},
		"high": {
			tt: tally{0, 17, 0, 0, 0, 0, 0},
			b:  scorer(0x40),
		},
		"medium": {
			tt: tally{0, 0, 5, 0, 0, 0, 0},
			b:  scorer(0x20),
		},
		"low": {
			tt: tally{0, 0, 0, 10, 0, 0, 0},
			b:  scorer(0x10),
		},
		"negligible": {
			tt: tally{0, 0, 0, 0, 10, 0, 0},
			b:  scorer(0x08),
		},
		"unknown": {
			tt: tally{0, 0, 0, 0, 0, 10, 0},
			b:  scorer(0x04),
		},
		"c/h": {
			tt: tally{10, 20, 0, 0, 0, 0, 0},
			b:  scorer(0xC0),
		},
		"c/m": {
			tt: tally{10, 0, 20, 0, 0, 0, 0},
			b:  scorer(0xA0),
		},
		"c/h/l": {
			tt: tally{10, 1, 20, 0, 0, 0, 0},
			b:  scorer(0xE0),
		},
		"n/u": {
			tt: tally{0, 0, 0, 0, 10, 20, 0},
			b:  scorer(0x0C),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.b, newScorer(u.tt))
		})
	}
}
