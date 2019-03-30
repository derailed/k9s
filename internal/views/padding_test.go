package views

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/stretchr/testify/assert"
)

func TestMaxColumn(t *testing.T) {
	uu := []struct {
		t resource.TableData
		s int
		e maxyPad
	}{
		{
			resource.TableData{
				Header: resource.Row{"A", "B"},
				Rows: resource.RowEvents{
					"r1": &resource.RowEvent{Fields: resource.Row{"hello", "world"}},
					"r2": &resource.RowEvent{Fields: resource.Row{"yo", "mama"}},
				},
			},
			0,
			maxyPad{5, 5},
		},
		{
			resource.TableData{
				Header: resource.Row{"A", "B"},
				Rows: resource.RowEvents{
					"r1": &resource.RowEvent{Fields: resource.Row{"hello", "world"}},
					"r2": &resource.RowEvent{Fields: resource.Row{"yo", "mama"}},
				},
			},
			1,
			maxyPad{5, 5},
		},
		{
			resource.TableData{
				Header: resource.Row{"A", "B"},
				Rows: resource.RowEvents{
					"r1": &resource.RowEvent{Fields: resource.Row{"Hello World lord of ipsums 😅", "world"}},
					"r2": &resource.RowEvent{Fields: resource.Row{"o", "mama"}},
				},
			},
			0,
			maxyPad{28, 5},
		},
	}

	for _, u := range uu {
		pads := make(maxyPad, len(u.t.Header))
		computeMaxColumns(pads, u.s, u.t)
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
		assert.Equal(t, u.e, isASCII(u.s))
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
		assert.Equal(t, u.e, pad(u.s, u.l))
	}
}

func BenchmarkMaxColumn(b *testing.B) {
	table := resource.TableData{
		Header: resource.Row{"A", "B"},
		Rows: resource.RowEvents{
			"r1": &resource.RowEvent{Fields: resource.Row{"hello", "world"}},
			"r2": &resource.RowEvent{Fields: resource.Row{"yo", "mama"}},
		},
	}

	pads := make(maxyPad, len(table.Header))

	b.ReportAllocs()
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		computeMaxColumns(pads, 0, table)
	}
}
