// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import (
	"fmt"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSortLabels(t *testing.T) {
	uu := map[string]struct {
		labels string
		e      [][]string
	}{
		"simple": {
			labels: "a=b,c=d",
			e: [][]string{
				{"a", "c"},
				{"b", "d"},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			hh, vv := sortLabels(labelize(u.labels))
			assert.Equal(t, u.e[0], hh)
			assert.Equal(t, u.e[1], vv)
		})
	}
}

func TestLabelize(t *testing.T) {
	uu := map[string]struct {
		labels string
		e      map[string]string
	}{
		"simple": {
			labels: "a=b,c=d",
			e:      map[string]string{"a": "b", "c": "d"},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, labelize(u.labels))
		})
	}
}

func TestIsValid(t *testing.T) {
	uu := map[string]struct {
		ns string
		h  Header
		r  Row
		e  bool
	}{
		"empty": {
			ns: "blee",
			h:  Header{},
			r:  Row{},
			e:  true,
		},
		"valid": {
			ns: "blee",
			h:  Header{HeaderColumn{Name: "VALID"}},
			r:  Row{Fields: Fields{"true"}},
			e:  true,
		},
		"invalid": {
			ns: "blee",
			h:  Header{HeaderColumn{Name: "VALID"}},
			r:  Row{Fields: Fields{"false"}},
			e:  false,
		},
		"valid_capital_case": {
			ns: "blee",
			h:  Header{HeaderColumn{Name: "VALID"}},
			r:  Row{Fields: Fields{"True"}},
			e:  true,
		},
		"valid_all_caps": {
			ns: "blee",
			h:  Header{HeaderColumn{Name: "VALID"}},
			r:  Row{Fields: Fields{"TRUE"}},
			e:  true,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			valid := IsValid(u.ns, u.h, u.r)
			assert.Equal(t, u.e, valid)
		})
	}
}

func TestDurationToSecond(t *testing.T) {
	uu := map[string]struct {
		s string
		e int64
	}{
		"seconds":                 {s: "22s", e: 22},
		"minutes":                 {s: "22m", e: 1320},
		"hours":                   {s: "12h", e: 43200},
		"days":                    {s: "3d", e: 259200},
		"day_hour":                {s: "3d9h", e: 291600},
		"day_hour_minute":         {s: "2d22h3m", e: 252180},
		"day_hour_minute_seconds": {s: "2d22h3m50s", e: 252230},
		"year":                    {s: "3y", e: 94608000},
		"year_day":                {s: "1y2d", e: 31708800},
		"na":                      {s: NAValue, e: math.MaxInt64},
		"n/a":                     {s: "n/a", e: math.MaxInt64},
		"unknown":                 {s: "<unknown>", e: math.MaxInt64},
		"none":                    {s: "<none>", e: math.MaxInt64},
		"invalid":                 {s: "<invalid>", e: math.MaxInt64},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, durationToSeconds(u.s))
		})
	}
}

func TestCapacityToQuantity(t *testing.T) {
	uu := map[string]struct {
		s  string
		e  string
		ok bool
	}{
		"empty":  {s: "", e: "0", ok: true},
		"blank":  {s: "  ", e: "0", ok: true},
		"1Gi":    {s: "1Gi", e: "1Gi", ok: true},
		"10Mi":   {s: "10Mi", e: "10Mi", ok: true},
		"millis": {s: "250m", e: "250m", ok: true},
		"dec":    {s: "0.5", e: "500m", ok: true},
		"na":     {s: "n/a"},
		"none":   {s: "<none>"},
		"ratio":  {s: "1/3"},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			q, ok := capacityToQuantity(u.s)
			assert.Equal(t, u.ok, ok)
			if u.ok {
				assert.Equal(t, u.e, q.String())
			}
		})
	}
}

func TestParallelRender(t *testing.T) {
	uu := map[string]struct {
		n int
	}{
		"empty":  {n: 0},
		"one":    {n: 1},
		"small":  {n: 7},
		"medium": {n: 100},
		"large":  {n: 10_000},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			results := make([]int, u.n)
			err := parallelRender(u.n, func(i int) error {
				results[i] = i + 1
				return nil
			})
			require.NoError(t, err)
			for i := range u.n {
				assert.Equal(t, i+1, results[i], "index %d", i)
			}
		})
	}
}

func TestParallelRenderError(t *testing.T) {
	err := parallelRender(50, func(i int) error {
		if i == 25 {
			return fmt.Errorf("boom at %d", i)
		}
		return nil
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "boom")
}

func BenchmarkParallelRender(b *testing.B) {
	const n = 50_000
	results := make([]int, n)

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		_ = parallelRender(n, func(i int) error {
			results[i] = i
			return nil
		})
	}
}

func BenchmarkDurationToSecond(b *testing.B) {
	t := "2d22h3m50s"

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		durationToSeconds(t)
	}
}
