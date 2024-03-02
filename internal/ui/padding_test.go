// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/stretchr/testify/assert"
)

func TestMaxColumn(t *testing.T) {
	uu := map[string]struct {
		t *model1.TableData
		s string
		e MaxyPad
	}{
		"ascii col 0": {
			model1.NewTableDataWithRows(
				client.NewGVR("test"),
				model1.Header{model1.HeaderColumn{Name: "A"}, model1.HeaderColumn{Name: "B"}},
				model1.NewRowEventsWithEvts(
					model1.RowEvent{
						Row: model1.Row{
							Fields: model1.Fields{"hello", "world"},
						},
					},
					model1.RowEvent{
						Row: model1.Row{
							Fields: model1.Fields{"yo", "mama"},
						},
					},
				),
			),
			"A",
			MaxyPad{6, 6},
		},
		"ascii col 1": {
			model1.NewTableDataWithRows(
				client.NewGVR("test"),
				model1.Header{model1.HeaderColumn{Name: "A"}, model1.HeaderColumn{Name: "B"}},
				model1.NewRowEventsWithEvts(
					model1.RowEvent{
						Row: model1.Row{
							Fields: model1.Fields{"hello", "world"},
						},
					},
					model1.RowEvent{
						Row: model1.Row{
							Fields: model1.Fields{"yo", "mama"},
						},
					},
				),
			),
			"B",
			MaxyPad{6, 6},
		},
		"non_ascii": {
			model1.NewTableDataWithRows(
				client.NewGVR("test"),
				model1.Header{model1.HeaderColumn{Name: "A"}, model1.HeaderColumn{Name: "B"}},
				model1.NewRowEventsWithEvts(
					model1.RowEvent{
						Row: model1.Row{
							Fields: model1.Fields{"Hello World lord of ipsums 😅", "world"},
						},
					},
					model1.RowEvent{
						Row: model1.Row{
							Fields: model1.Fields{"o", "mama"},
						},
					},
				),
			),
			"A",
			MaxyPad{32, 6},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			pads := make(MaxyPad, u.t.HeaderCount())
			ComputeMaxColumns(pads, u.s, u.t)
			assert.Equal(t, u.e, pads)
		})
	}
}

func TestIsASCII(t *testing.T) {
	uu := []struct {
		s string
		e bool
	}{
		{"hello", true},
		{"Yo! 😄", false},
		{"😄", false},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, IsASCII(u.s))
	}
}

func TestPad(t *testing.T) {
	uu := []struct {
		s string
		l int
		e string
	}{
		{"fred", 3, "fr…"},
		{"01234567890", 10, "012345678…"},
		{"fred", 10, "fred      "},
		{"fred", 6, "fred  "},
		{"fred", 4, "fred"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, Pad(u.s, u.l))
	}
}

func BenchmarkMaxColumn(b *testing.B) {
	table := model1.NewTableDataWithRows(
		client.NewGVR("test"),
		model1.Header{model1.HeaderColumn{Name: "A"}, model1.HeaderColumn{Name: "B"}},
		model1.NewRowEventsWithEvts(
			model1.RowEvent{
				Row: model1.Row{
					Fields: model1.Fields{"hello", "world"},
				},
			},
			model1.RowEvent{
				Row: model1.Row{
					Fields: model1.Fields{"yo", "mama"},
				},
			},
		),
	)

	pads := make(MaxyPad, table.HeaderCount())

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ComputeMaxColumns(pads, "A", table)
	}
}
