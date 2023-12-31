// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package vul

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_newTally(t *testing.T) {
	uu := map[string]struct {
		t  *table
		tt tally
	}{
		"full": {
			t:  makeTable(t, "testdata/sort/full/sc2.text"),
			tt: tally{7, 14, 8, 0, 0, 0, 29},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.tt, newTally(u.t))
		})
	}
}

func Test_score(t *testing.T) {
	uu := map[string]struct {
		tt tally
		sc int
	}{
		"zero": {},
		"critical": {
			tt: tally{29, 7, 14, 8, 0, 0, 0},
			sc: 292180,
		},
		"high": {
			tt: tally{0, 17, 14, 8, 0, 0, 0},
			sc: 3180,
		},
		"medium": {
			tt: tally{0, 0, 14, 0, 0, 0, 0},
			sc: 1400,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.sc, u.tt.score())
		})
	}
}
