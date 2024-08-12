// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package vul

import (
	"fmt"
	"io"
)

const (
	sevCritical = iota
	sevHigh
	sevMedium
	sevLow
	sevNegligible
	sevUnknown
	sevFixed
)

var vulWeights = []int{10_000, 100, 100, 10, 0, 0, 0, 0}

type tally [7]int

func newTally(t *table) tally {
	var tt tally
	for _, r := range t.Rows {
		if r.Fix() != "" {
			tt[sevFixed]++
		}
		switch r.Severity() {
		case Sev1:
			tt[sevCritical]++
		case Sev2:
			tt[sevHigh]++
		case Sev3:
			tt[sevMedium]++
		case Sev4:
			tt[sevLow]++
		case Sev5:
			tt[sevNegligible]++
		case SevU:
			tt[sevUnknown]++
		}
	}

	return tt
}

// Dump dumps tally as text.
func (t tally) Dump(w io.Writer) {
	fmt.Fprintf(w, "%d critical, %d high, %d medium, %d low, %d negligible",
		t[sevCritical],
		t[sevHigh],
		t[sevMedium],
		t[sevLow],
		t[sevNegligible],
	)
	if t[sevUnknown] > 0 {
		fmt.Fprintf(w, " (%d unknown)", t[sevUnknown])
	}
	if t[sevFixed] > 0 {
		fmt.Fprintf(w, " -- [Fixed: %d]", t[sevFixed])
	}
}

func (t *tally) score() int {
	var s int
	for i, v := range t[:5] {
		s += v * vulWeights[i]
	}

	return s
}
