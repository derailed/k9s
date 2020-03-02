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
			m: Metric{OK: 100, Fault: 10},
			s: 0.5,
			e: blocks{
				oks:  block{full: 6, partial: sparks[2]},
				errs: block{full: 0, partial: sparks[5]},
			},
		},
		"max_fault": {
			m: Metric{OK: 10, Fault: 100},
			s: 0.5,
			e: blocks{
				oks:  block{full: 0, partial: sparks[5]},
				errs: block{full: 6, partial: sparks[2]},
			},
		},
		"over": {
			m: Metric{OK: 22, Fault: 999},
			s: float64(8*20) / float64(999),
			e: blocks{
				oks:  block{full: 0, partial: sparks[4]},
				errs: block{full: 20, partial: sparks[0]},
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
		e  int
	}{
		"empty": {
			e: 0,
		},
		"max_ok": {
			mm: []Metric{{OK: 100, Fault: 10}},
			e:  100,
		},
		"max_fault": {
			mm: []Metric{{OK: 100, Fault: 1000}},
			e:  1000,
		},
		"many": {
			mm: []Metric{
				{OK: 100, Fault: 1000},
				{OK: 110, Fault: 1010},
				{OK: 120, Fault: 1020},
				{OK: 130, Fault: 1030},
				{OK: 140, Fault: 1040},
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
