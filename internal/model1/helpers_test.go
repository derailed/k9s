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
	for range b.N {
		durationToSeconds(t)
	}
}

func TestIsDurationFormat(t *testing.T) {
	uu := map[string]struct {
		s        string
		expected bool
	}{
		"empty string":              {s: "", expected: false},
		"na value":                  {s: NAValue, expected: false},
		"simple seconds":            {s: "30s", expected: true},
		"simple minutes":            {s: "15m", expected: true},
		"simple hours":              {s: "2h", expected: true},
		"simple days":               {s: "7d", expected: true},
		"simple years":              {s: "1y", expected: true},
		"combined duration":         {s: "1h30m5s", expected: true},
		"full mixed duration":       {s: "2y3d4h5m6s", expected: true},
		"missing unit after number": {s: "30", expected: false},
		"letters without numbers":   {s: "hms", expected: false},
		"invalid characters":        {s: "30x", expected: false},
		"multiple parts (space)":    {s: "1h 30m", expected: false},
		"starting with letter":      {s: "h30m", expected: false},
		"ending with number":        {s: "30m5", expected: false},
		"only number":               {s: "123", expected: false},
		"unit without number":       {s: "h", expected: false},
		"zero seconds":              {s: "0s", expected: true},
		"multiple zeros":            {s: "0h0m0s", expected: true},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.expected, isDurationFormat(u.s))
		})
	}
}
