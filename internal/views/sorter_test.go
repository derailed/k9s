package views

import (
	"sort"
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestGroupSort(t *testing.T) {
	uu := []struct {
		asc    bool
		rows   []string
		expect []string
	}{
		{true, []string{"200m", "100m"}, []string{"100m", "200m"}},
		{true, []string{"200m", "100m"}, []string{"100m", "200m"}},
		{false, []string{"200m", "100m"}, []string{"200m", "100m"}},
		{true, []string{"10", "1"}, []string{"1", "10"}},
		{false, []string{"10", "1"}, []string{"10", "1"}},
		{true, []string{"100Mi", "10Mi"}, []string{"10Mi", "100Mi"}},
		{false, []string{"100Mi", "10Mi"}, []string{"100Mi", "10Mi"}},
		{true, []string{"xyz", "abc"}, []string{"abc", "xyz"}},
		{false, []string{"xyz", "abc"}, []string{"xyz", "abc"}},
		{true, []string{"2m30s", "1m10s"}, []string{"1m10s", "2m30s"}},
		{true, []string{"3d", "1d"}, []string{"1d", "3d"}},

		{true, []string{"95h", "93h"}, []string{"93h", "95h"}},
		{true, []string{"95d", "93d"}, []string{"93d", "95d"}},
		{true, []string{"1h10m", "59m"}, []string{"59m", "1h10m"}},
		{true, []string{"95m", "1h30m"}, []string{"1h30m", "95m"}},
		{true, []string{"b-21", "b-2"}, []string{"b-2", "b-21"}},
		{false, []string{"b-21", "b-2"}, []string{"b-21", "b-2"}},
		{true, []string{"4m", "3m2s"}, []string{"3m2s", "4m"}},
		{true, []string{"3y37d", "2y4d"}, []string{"2y4d", "3y37d"}},
	}

	for _, u := range uu {
		g := groupSorter{rows: u.rows, asc: u.asc}
		sort.Sort(g)
		assert.Equal(t, u.expect, g.rows)
	}
}

func TestRowSort(t *testing.T) {
	uu := []struct {
		asc          bool
		rows, expect resource.Rows
	}{
		{
			true,
			resource.Rows{resource.Row{"200m"}, resource.Row{"100m"}},
			resource.Rows{resource.Row{"100m"}, resource.Row{"200m"}},
		},
		{
			false,
			resource.Rows{resource.Row{"200m"}, resource.Row{"100m"}},
			resource.Rows{resource.Row{"200m"}, resource.Row{"100m"}},
		},
		{
			true,
			resource.Rows{resource.Row{"200Mi"}, resource.Row{"100Mi"}},
			resource.Rows{resource.Row{"100Mi"}, resource.Row{"200Mi"}},
		},
		{
			false,
			resource.Rows{resource.Row{"200Mi"}, resource.Row{"100Mi"}},
			resource.Rows{resource.Row{"200Mi"}, resource.Row{"100Mi"}},
		},
		{
			true,
			resource.Rows{resource.Row{"8m4s"}, resource.Row{"31m"}},
			resource.Rows{resource.Row{"8m4s"}, resource.Row{"31m"}},
		},
		{
			true,
			resource.Rows{resource.Row{"n/a"}, resource.Row{"31m"}},
			resource.Rows{resource.Row{"31m"}, resource.Row{"n/a"}},
		},
		{
			true,
			resource.Rows{resource.Row{"31m"}, resource.Row{"n/a"}},
			resource.Rows{resource.Row{"31m"}, resource.Row{"n/a"}},
		},
		{
			false,
			resource.Rows{resource.Row{"n/a"}, resource.Row{"31m"}},
			resource.Rows{resource.Row{"31m"}, resource.Row{"n/a"}},
		},
		{
			false,
			resource.Rows{resource.Row{"31m"}, resource.Row{"n/a"}},
			resource.Rows{resource.Row{"31m"}, resource.Row{"n/a"}},
		},
	}

	for _, u := range uu {
		r := rowSorter{index: 0, rows: u.rows, asc: u.asc}
		sort.Sort(r)
		assert.Equal(t, u.expect, r.rows)
	}
}

func TestIsDurationSort(t *testing.T) {
	uu := map[string]struct {
		s1, s2 string
		asc, e bool
	}{
		"ascLess":     {"10h5m", "2h10m", true, false},
		"descGreater": {"10h5m", "2h10m", false, true},
		"ascEqual":    {"2h10m", "2h10m", true, true},
		"descEqual":   {"2h10m", "2h10m", false, true},
		"ascGreater":  {"10h10m", "2h5m", true, false},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			less, ok := isDurationSort(u.asc, u.s1, u.s2)
			assert.True(t, ok)
			assert.Equal(t, u.e, less)
		})
	}
}
