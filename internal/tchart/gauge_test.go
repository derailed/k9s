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
			m: tchart.Metric{OK: 100, Fault: 10},
			e: 3,
		},
		"errs": {
			m: tchart.Metric{OK: 10, Fault: 1000},
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
		e int
	}{
		"empty": {
			e: 0,
		},
		"max_ok": {
			m: tchart.Metric{OK: 100, Fault: 10},
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

func TestGauge(t *testing.T) {
	uu := map[string]struct {
		mm []tchart.Metric
		e  int
	}{
		"empty": {
			e: 1,
		},
		"oks": {
			mm: []tchart.Metric{{OK: 100, Fault: 10}},
			e:  3,
		},
		"errs": {
			mm: []tchart.Metric{{OK: 10, Fault: 1000}},
			e:  4,
		},
	}

	for k := range uu {
		u := uu[k]
		g := tchart.NewGauge("fred")
		assert.True(t, g.IsDial())
		for _, m := range u.mm {
			g.Add(m)
		}
		t.Run(k, func(t *testing.T) {
			// assert.Equal(t, u.e, u.m.MaxDigits())
		})
	}

}
