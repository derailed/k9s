package ui

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestMaxColumn(t *testing.T) {
	uu := map[string]struct {
		t resource.TableData
		s int
		e MaxyPad
	}{
		"ascii col 0": {
			resource.TableData{
				Header: render.HeaderRow{render.Header{Name: "A"}, render.Header{Name: "B"}},
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
			0,
			MaxyPad{6, 6},
		},
		"ascii col 1": {
			resource.TableData{
				Header: render.HeaderRow{render.Header{Name: "A"}, render.Header{Name: "B"}},
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
			1,
			MaxyPad{6, 6},
		},
		"non_ascii": {
			resource.TableData{
				Header: render.HeaderRow{render.Header{Name: "A"}, render.Header{Name: "B"}},
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
			0,
			MaxyPad{32, 6},
		},
	}

	for _, u := range uu {
		pads := make(MaxyPad, len(u.t.Header))
		ComputeMaxColumns(pads, u.s, u.t.Header, u.t.RowEvents)
		assert.Equal(t, u.e, pads)
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
	table := resource.TableData{
		Header: render.HeaderRow{render.Header{Name: "A"}, render.Header{Name: "B"}},
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
		ComputeMaxColumns(pads, 0, table.Header, table.RowEvents)
	}
}
