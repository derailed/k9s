// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package tchart_test

import (
	"testing"

	"github.com/derailed/k9s/internal/tchart"
	"github.com/stretchr/testify/assert"
)

func TestMetricsMaxDigits(t *testing.T) {
	uu := map[string]struct {
		m tchart.Metric
		e int
	}{
		"empty": {
			e: 1,
		},
		"oks": {
			m: tchart.Metric{S1: 100, S2: 10},
			e: 3,
		},
		"errs": {
			m: tchart.Metric{S1: 10, S2: 1000},
			e: 4,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.m.MaxDigits())
		})
	}
}

func TestMetricsMax(t *testing.T) {
	uu := map[string]struct {
		m tchart.Metric
		e int64
	}{
		"empty": {
			e: 0,
		},
		"max_ok": {
			m: tchart.Metric{S1: 100, S2: 10},
			e: 100,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.m.Max())
		})
	}
}
