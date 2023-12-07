// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package vul

import "fmt"

type scorer uint8

func (b scorer) String() string {
	return fmt.Sprintf("%08b", b)[:6]
}

func newScorer(t tally) scorer {
	return fromTally(t)
}

func (b scorer) Add(b1 scorer) scorer {
	return b | b1
}

func fromTally(t tally) scorer {
	var b scorer
	for i, v := range t {
		if v == 0 {
			continue
		}
		switch i {
		case sevCritical:
			b |= 0x80
		case sevHigh:
			b |= 0x40
		case sevMedium:
			b |= 0x20
		case sevLow:
			b |= 0x10
		case sevNegligible:
			b |= 0x08
		case sevUnknown:
			b |= 0x04
		}
	}

	return b
}
