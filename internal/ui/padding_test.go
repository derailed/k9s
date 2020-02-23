package ui

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestMaxColumn(t *testing.T) {
	uu := map[string]struct {
		t render.TableData
		s string
		e MaxyPad
	}{
		"ascii col 0": {
			render.TableData{
				Header: render.Header{render.HeaderColumn{Name: "A"}, render.HeaderColumn{Name: "B"}},
				RowEvents: render.RowEvents{
					render.RowEvent{
						Row: render.Row{
							Fields: render.Fields{"hello", "world"},
						},
					},
					render.RowEvent{
						Row: render.Row{
							Fields: render.Fields{"yo", "mama"},
						},
					},
				},
			},
			"A",
			MaxyPad{6, 6},
		},
		"ascii col 1": {
			render.TableData{
				Header: render.Header{render.HeaderColumn{Name: "A"}, render.HeaderColumn{Name: "B"}},
				RowEvents: render.RowEvents{
					render.RowEvent{
						Row: render.Row{
							Fields: render.Fields{"hello", "world"},
						},
					},
					render.RowEvent{
						Row: render.Row{
							Fields: render.Fields{"yo", "mama"},
						},
					},
				},
			},
			"B",
			MaxyPad{6, 6},
		},
		"non_ascii": {
			render.TableData{
				Header: render.Header{render.HeaderColumn{Name: "A"}, render.HeaderColumn{Name: "B"}},
				RowEvents: render.RowEvents{
					render.RowEvent{
						Row: render.Row{
							Fields: render.Fields{"Hello World lord of ipsums 😅", "world"},
						},
					},
					render.RowEvent{
						Row: render.Row{
							Fields: render.Fields{"o", "mama"},
						},
					},
				},
			},
			"A",
			MaxyPad{32, 6},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			pads := make(MaxyPad, len(u.t.Header))
			ComputeMaxColumns(pads, u.s, u.t.Header, u.t.RowEvents)
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
	table := render.TableData{
		Header: render.Header{render.HeaderColumn{Name: "A"}, render.HeaderColumn{Name: "B"}},
		RowEvents: render.RowEvents{
			render.RowEvent{
				Row: render.Row{
					Fields: render.Fields{"hello", "world"},
				},
			},
			render.RowEvent{
				Row: render.Row{
					Fields: render.Fields{"yo", "mama"},
				},
			},
		},
	}

	pads := make(MaxyPad, len(table.Header))

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		ComputeMaxColumns(pads, "A", table.Header, table.RowEvents)
	}
}
