// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
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
		"n/a":                     {s: NAValue, e: math.MaxInt64},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, durationToSeconds(u.s))
		})
	}
}

func BenchmarkDurationToSecond(b *testing.B) {
	t := "2d22h3m50s"

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		durationToSeconds(t)
	}
}
