package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestRowDelete(t *testing.T) {
	uu := map[string]struct {
		rows render.Rows
		id   string
		e    render.Rows
	}{
		"first": {
			rows: render.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
				{ID: "b", Fields: []string{"albert", "blee"}},
			},
			id: "a",
			e: render.Rows{
				{ID: "b", Fields: []string{"albert", "blee"}},
			},
		},
		"last": {
			rows: render.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
				{ID: "b", Fields: []string{"albert", "blee"}},
			},
			id: "b",
			e: render.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
			},
		},
		"middle": {
			rows: render.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
				{ID: "b", Fields: []string{"albert", "blee"}},
				{ID: "c", Fields: []string{"fred", "zorg"}},
			},
			id: "b",
			e: render.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
				{ID: "c", Fields: []string{"fred", "zorg"}},
			},
		},
		"missing": {
			rows: render.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
				{ID: "b", Fields: []string{"albert", "blee"}},
			},
			id: "zorg",
			e: render.Rows{
				{ID: "a", Fields: []string{"blee", "duh"}},
				{ID: "b", Fields: []string{"albert", "blee"}},
			},
		},
	}

	for k := range uu {
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			rows := uc.rows.Delete(uc.id)
			assert.Equal(t, uc.e, rows)
		})
	}
}

func TestSortText(t *testing.T) {
	uu := map[string]struct {
		rows render.Rows
		col  int
		asc  bool
		e    render.Rows
	}{
		"plainAsc": {
			rows: render.Rows{
				{Fields: []string{"blee", "duh"}},
				{Fields: []string{"albert", "blee"}},
			},
			col: 0,
			asc: true,
			e: render.Rows{
				{Fields: []string{"albert", "blee"}},
				{Fields: []string{"blee", "duh"}},
			},
		},
		"plainDesc": {
			rows: render.Rows{
				{Fields: []string{"blee", "duh"}},
				{Fields: []string{"albert", "blee"}},
			},
			col: 0,
			asc: false,
			e: render.Rows{
				{Fields: []string{"blee", "duh"}},
				{Fields: []string{"albert", "blee"}},
			},
		},
		"numericAsc": {
			rows: render.Rows{
				{Fields: []string{"10", "duh"}},
				{Fields: []string{"1", "blee"}},
			},
			col: 0,
			asc: true,
			e: render.Rows{
				{Fields: []string{"1", "blee"}},
				{Fields: []string{"10", "duh"}},
			},
		},
		"numericDesc": {
			rows: render.Rows{
				{Fields: []string{"10", "duh"}},
				{Fields: []string{"1", "blee"}},
			},
			col: 0,
			asc: false,
			e: render.Rows{
				{Fields: []string{"10", "duh"}},
				{Fields: []string{"1", "blee"}},
			},
		},
		"composite": {
			rows: render.Rows{
				{Fields: []string{"blee-duh", "duh"}},
				{Fields: []string{"blee", "blee"}},
			},
			col: 0,
			asc: true,
			e: render.Rows{
				{Fields: []string{"blee", "blee"}},
				{Fields: []string{"blee-duh", "duh"}},
			},
		},
	}

	for k := range uu {
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			uc.rows.Sort(uc.col, uc.asc)
			assert.Equal(t, uc.e, uc.rows)
		})
	}
}

func TestSortDuration(t *testing.T) {
	uu := map[string]struct {
		rows render.Rows
		col  int
		asc  bool
		e    render.Rows
	}{
		"durationAsc": {
			rows: render.Rows{
				{Fields: []string{"10m10s", "duh"}},
				{Fields: []string{"19s", "blee"}},
			},
			col: 0,
			asc: true,
			e: render.Rows{
				{Fields: []string{"19s", "blee"}},
				{Fields: []string{"10m10s", "duh"}},
			},
		},
		"durationDesc": {
			rows: render.Rows{
				{Fields: []string{"10m10s", "duh"}},
				{Fields: []string{"19s", "blee"}},
			},
			col: 0,
			e: render.Rows{
				{Fields: []string{"10m10s", "duh"}},
				{Fields: []string{"19s", "blee"}},
			},
		},
	}

	for k := range uu {
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			uc.rows.Sort(uc.col, uc.asc)
			assert.Equal(t, uc.e, uc.rows)
		})
	}
}

func TestSortMetrics(t *testing.T) {
	uu := map[string]struct {
		rows render.Rows
		col  int
		asc  bool
		e    render.Rows
	}{
		"metricAsc": {
			rows: render.Rows{
				{Fields: []string{"10m", "duh"}},
				{Fields: []string{"1m", "blee"}},
			},
			col: 0,
			asc: true,
			e: render.Rows{
				{Fields: []string{"1m", "blee"}},
				{Fields: []string{"10m", "duh"}},
			},
		},
		"metricDesc": {
			rows: render.Rows{
				{Fields: []string{"10m", "100Mi"}},
				{Fields: []string{"1m", "50Mi"}},
			},
			col: 1,
			asc: false,
			e: render.Rows{
				{Fields: []string{"10m", "100Mi"}},
				{Fields: []string{"1m", "50Mi"}},
			},
		},
	}

	for k := range uu {
		uc := uu[k]
		t.Run(k, func(t *testing.T) {
			uc.rows.Sort(uc.col, uc.asc)
			assert.Equal(t, uc.e, uc.rows)
		})
	}
}
