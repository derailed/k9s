package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNamespaced(t *testing.T) {
	uu := []struct {
		p, ns, n string
	}{
		{"fred/blee", "fred", "blee"},
	}

	for _, u := range uu {
		ns, n := namespaced(u.p)
		assert.Equal(t, u.ns, ns)
		assert.Equal(t, u.n, n)
	}
}

func TestMissing(t *testing.T) {
	uu := []struct {
		i, e string
	}{
		{"fred", "fred"},
		{"", MissingValue},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, missing(u.i))
	}
}

func TestBoolToStr(t *testing.T) {
	uu := []struct {
		i bool
		e string
	}{
		{true, "true"},
		{false, "false"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, boolToStr(u.i))
	}
}

func TestNa(t *testing.T) {
	uu := []struct {
		i, e string
	}{
		{"fred", "fred"},
		{"", NAValue},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, na(u.i))
	}
}

func TestPad(t *testing.T) {
	uu := []struct {
		s string
		l int
		e string
	}{
		{"fred", 10, "fred      "},
		{"fred", 6, "fred  "},
		{"fred", 4, "fred"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, Pad(u.s, u.l))
	}
}

func TestTruncate(t *testing.T) {
	uu := []struct {
		s string
		l int
		e string
	}{
		{"fred", 2, "fr…"},
		{"fred", 1, "f…"},
		{"fred", 10, "fred"},
	}

	for _, u := range uu {
		assert.Equal(t, u.e, Truncate(u.s, u.l))
	}
}

func TestMapToStr(t *testing.T) {
	uu := []struct {
		i map[string]string
		e string
	}{
		{map[string]string{"blee": "duh", "aa": "bb"}, "aa=bb,blee=duh"},
		{map[string]string{}, MissingValue},
	}
	for _, u := range uu {
		assert.Equal(t, u.e, mapToStr(u.i))
	}
}

func BenchmarkMapToStr(b *testing.B) {
	ll := map[string]string{
		"blee": "duh",
		"aa":   "bb",
	}
	b.ResetTimer()
	b.ReportAllocs()

	for i := 0; i < b.N; i++ {
		mapToStr(ll)
	}
}
