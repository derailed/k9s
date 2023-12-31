// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tchart

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCutSet(t *testing.T) {
	uu := map[string]struct {
		mm   []Metric
		w, e int
	}{
		"empty": {
			w: 10,
			e: 0,
		},
		"at": {
			mm: make([]Metric, 10),
			w:  10,
			e:  10,
		},
		"under": {
			mm: make([]Metric, 5),
			w:  10,
			e:  5,
		},
		"over": {
			mm: make([]Metric, 10),
			w:  5,
			e:  5,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			s := NewSparkLine("s")
			assert.False(t, s.IsDial())

			for _, m := range u.mm {
				s.Add(m)
			}
			s.cutSet(u.w)
			assert.Equal(t, u.e, len(s.data))
		})
	}
}

func TestToBlocks(t *testing.T) {
	uu := map[string]struct {
		m Metric
		s float64
		e blocks
	}{
		"empty": {
			e: blocks{},
		},
		"max_ok": {
			m: Metric{S1: 100, S2: 10},
			s: 0.5,
			e: blocks{
				s1: block{full: 6, partial: sparks[2]},
				s2: block{full: 0, partial: sparks[5]},
			},
		},
		"max_fault": {
			m: Metric{S1: 10, S2: 100},
			s: 0.5,
			e: blocks{
				s1: block{full: 0, partial: sparks[5]},
				s2: block{full: 6, partial: sparks[2]},
			},
		},
		"over": {
			m: Metric{S1: 22, S2: 999},
			s: float64(8*20) / float64(999),
			e: blocks{
				s1: block{full: 0, partial: sparks[4]},
				s2: block{full: 20, partial: sparks[0]},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, toBlocks(u.m, u.s))
		})
	}
}

func TestComputeMax(t *testing.T) {
	uu := map[string]struct {
		mm []Metric
		e  int64
	}{
		"empty": {
			e: 0,
		},
		"max_ok": {
			mm: []Metric{{S1: 100, S2: 10}},
			e:  100,
		},
		"max_fault": {
			mm: []Metric{{S1: 100, S2: 1000}},
			e:  1000,
		},
		"many": {
			mm: []Metric{
				{S1: 100, S2: 1000},
				{S1: 110, S2: 1010},
				{S1: 120, S2: 1020},
				{S1: 130, S2: 1030},
				{S1: 140, S2: 1040},
			},
			e: 1040,
		},
	}

	for k := range uu {
		u := uu[k]
		s := NewSparkLine("s")
		for _, m := range u.mm {
			s.Add(m)
		}
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, s.computeMax())
		})
	}
}
